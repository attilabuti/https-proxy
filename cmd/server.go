package cmd

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"encoding/base64"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unsafe"
)

type appServer struct {
	http   http.Server
	https  http.Server
	idle   chan struct{}
	errors chan error
}

const (
	BasicAuthName         = "Basic"
	ProxyAuthorizationKey = "Proxy-Authorization"
	ProxyAuthenticateKey  = "Proxy-Authenticate"
)

// Hop-by-hop headers. These are removed when sent to the backend.
// As of RFC 7230, hop-by-hop headers are required to appear in the
// Connection header field. These are the headers defined by the
// obsoleted RFC 2616 (section 13.5.1) and are used for backward
// compatibility.
var hopHeaders = []string{
	"Connection",
	"Proxy-Connection", // Non-standard but still sent by libcurl and rejected by e.g. google
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",      // Canonicalized version of "TE"
	"Trailer", // Not Trailers per URL above; https://www.rfc-editor.org/errata_search.php?eid=4522
	"Transfer-Encoding",
	"Upgrade",
}

func (s *appServer) start() {
	s.errors = make(chan error)
	s.idle = make(chan struct{})

	if config.http.enabled {
		s.http = s.createServer(config.http.address)

		go func() {
			log.info.Printf("HTTP server listening on %v\n", config.http.address)
			s.errors <- s.http.ListenAndServe()
		}()
	}

	if config.https.enabled {
		s.https = s.createServer(config.https.address)

		go func() {
			log.info.Printf("HTTPS server listening on %v\n", config.https.address)
			s.errors <- s.https.ListenAndServeTLS(config.https.cert, config.https.key)
		}()
	}

	go s.close()

	log.error.Println(<-s.errors)

	<-s.idle
}

func (s *appServer) close() {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-sigint
	s.stop()
}

func (s *appServer) stop() {
	if config.http.enabled {
		if err := s.http.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.error.Printf("HTTP server shutdown error: %v\n", err)
		} else {
			log.info.Println("HTTP server shutdown")
		}
	}

	if config.https.enabled {
		if err := s.https.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.error.Printf("HTTPS server shutdown error: %v\n", err)
		} else {
			log.info.Println("HTTPS server shutdown")
		}
	}

	log.close()

	close(s.errors)
	close(s.idle)
}

func (s *appServer) createServer(addr string) http.Server {
	return http.Server{
		Addr:         addr,
		ErrorLog:     log.error,
		ReadTimeout:  config.timeout.readDuration,
		WriteTimeout: config.timeout.writeDuration,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.auth.enabled {
				if !s.basicAuth(r.Header.Get(ProxyAuthorizationKey)) {
					log.error.Printf("authentication failed %v", r.RemoteAddr)

					w.Header().Set(ProxyAuthenticateKey, BasicAuthName)
					http.Error(w, http.StatusText(http.StatusProxyAuthRequired), http.StatusProxyAuthRequired)

					return
				}
			}

			log.conn.Printf("new connection from %v", r.RemoteAddr)

			if r.Method == http.MethodConnect {
				s.handleTunneling(w, r)
			} else {
				s.handleHTTP(w, r)
			}
		}),

		// Disable HTTP/2
		// https://github.com/golang/go/issues/14797#issuecomment-196103814
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
}

func (s *appServer) basicAuth(auth string) bool {
	if u, p, ok := s.parseBasicAuth(auth); ok {
		username := sha256.Sum256([]byte(u))
		password := sha256.Sum256([]byte(p))

		usernameMatch := (subtle.ConstantTimeCompare(config.auth.userHash[:], username[:]) == 1)
		passwordMatch := (subtle.ConstantTimeCompare(config.auth.passHash[:], password[:]) == 1)

		if usernameMatch && passwordMatch {
			return true
		}
	}

	return false
}

func (s *appServer) parseBasicAuth(auth string) (string, string, bool) {
	const prefix = BasicAuthName + " "

	if !strings.HasPrefix(auth, prefix) {
		return "", "", false
	}

	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return "", "", false
	}

	cs := *(*string)(unsafe.Pointer(&c))
	st := strings.IndexByte(cs, ':')
	if st < 0 {
		return "", "", false
	}

	return cs[:st], cs[st+1:], true
}

func (s *appServer) handleTunneling(w http.ResponseWriter, r *http.Request) {
	destConn, err := net.DialTimeout("tcp", r.Host, config.timeout.dialDuration)
	if err != nil {
		log.error.Printf("handleTunneling - DialTimeout error: %v\n", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		log.error.Printf("handleTunneling - Hijack error: %v\n", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	go s.transfer(destConn, clientConn)
	go s.transfer(clientConn, destConn)
}

func (s *appServer) transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()

	io.Copy(destination, source)
}

func (s *appServer) handleHTTP(w http.ResponseWriter, req *http.Request) {
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		log.error.Printf("handleHTTP - RoundTrip error: %v\n", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	defer resp.Body.Close()

	// Remove hop-by-hop headers
	for _, h := range hopHeaders {
		req.Header.Del(h)
	}

	// Copy headers
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)
}

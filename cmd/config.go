package cmd

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"
)

type configuration struct {
	host string

	http struct {
		enabled bool   // HTTP server enabled
		port    int    // HTTP server port
		address string // HTTP server address
	}

	https struct {
		enabled bool   // HTTPS server enabled
		port    int    // HTTPS server port
		address string // HTTPS server address
		cert    string // SSL certificate file
		key     string // RSA private key file
	}

	auth struct {
		enabled  bool // Proxy authentication enabled
		username string
		password string
		userHash [32]byte
		passHash [32]byte
	}

	timeout struct {
		read          int
		write         int
		dial          int
		readDuration  time.Duration
		writeDuration time.Duration
		dialDuration  time.Duration
	}

	log struct {
		enabled     bool   // Logging enabled
		dir         string // Log files directory
		connections bool   // Log HTTP(S) connections
	}

	quiet bool
}

func (c *configuration) init() error {
	var err error

	if !c.http.enabled && !c.https.enabled {
		return errors.New("HTTP or HTTPS must be enabled")
	}

	if c.http.enabled {
		if c.http.port == 0 {
			return errors.New("HTTP port must be specified")
		} else if c.http.port < 0 || c.http.port > 65535 {
			return fmt.Errorf("invalid HTTP port number: %v", c.http.port)
		}

		c.http.address = net.JoinHostPort(c.host, strconv.Itoa(c.http.port))
	}

	if c.https.enabled {
		if c.https.port == 0 {
			return errors.New("HTTPS port must be specified")
		} else if c.https.port < 0 || c.https.port > 65535 {
			return fmt.Errorf("invalid HTTPS port number: %v", c.https.port)
		}

		c.https.address = net.JoinHostPort(c.host, strconv.Itoa(c.https.port))

		if len(c.https.cert) == 0 {
			return errors.New("SSL certificate file must be specified")
		} else if !fileExists(c.https.cert) {
			return fmt.Errorf("SSL certificate file specified but not found: %s", c.https.cert)
		}

		if len(c.https.key) == 0 {
			return errors.New("RSA private key file must be specified")
		} else if !fileExists(c.https.key) {
			return fmt.Errorf("RSA private key file specified but not found: %s", c.https.key)
		}
	}

	if c.auth.enabled {
		if len(c.auth.username) > 0 {
			c.auth.userHash = sha256.Sum256([]byte(c.auth.username))
		} else {
			return errors.New("username cannot be empty")
		}

		if len(c.auth.password) > 0 {
			c.auth.passHash = sha256.Sum256([]byte(c.auth.password))
		} else {
			return errors.New("password cannot be empty")
		}
	}

	if c.timeout.read < 0 {
		return errors.New("timeout-read cannot be smaller than 0")
	} else {
		c.timeout.readDuration, err = time.ParseDuration(fmt.Sprintf("%vs", c.timeout.read))
		if err != nil {
			return fmt.Errorf("timeout-read: %s", err)
		}
	}

	if c.timeout.write < 0 {
		return errors.New("timeout-write cannot be smaller than 0")
	} else {
		c.timeout.writeDuration, err = time.ParseDuration(fmt.Sprintf("%vs", c.timeout.write))
		if err != nil {
			return fmt.Errorf("timeout-write: %s", err)
		}
	}

	if c.timeout.dial < 0 {
		return errors.New("timeout-dial cannot be smaller than 0")
	} else {
		c.timeout.dialDuration, err = time.ParseDuration(fmt.Sprintf("%vs", c.timeout.dial))
		if err != nil {
			return fmt.Errorf("timeout-dial: %s", err)
		}
	}

	return nil
}

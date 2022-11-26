package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

var (
	app struct {
		name string
		cli  *cli.App
		run  bool
	}
	server appServer
	config configuration
	log    logger
)

func Execute() {
	if err := app.cli.Run(os.Args); err != nil {
		fmt.Printf("%s: error: %s\n", app.name, err)
		fmt.Printf("Type %s --help to see a list of all options.", app.name)
		os.Exit(1)
	}

	if app.run {
		server.start()
	}
}

func init() {
	if len(os.Args) > 0 {
		app.name = filepath.Base(os.Args[0])
	}

	flags := []cli.Flag{
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "host",
			Value:       "",
			Usage:       "Server `host`",
			Destination: &config.host,
		}),

		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "enable-http",
			Usage:       "Enable HTTP server",
			Value:       false,
			Destination: &config.http.enabled,
		}),
		altsrc.NewIntFlag(&cli.IntFlag{
			Name:        "port-http",
			Value:       80,
			Usage:       "HTTP `port`",
			Destination: &config.http.port,
		}),

		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "enable-https",
			Usage:       "Enable HTTPS server",
			Value:       false,
			Destination: &config.https.enabled,
		}),
		altsrc.NewIntFlag(&cli.IntFlag{
			Name:        "port-https",
			Value:       443,
			Usage:       "HTTPS `port`",
			Destination: &config.https.port,
		}),

		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "crt-file",
			Usage:       "Location of the SSL certificate `file`",
			Value:       "",
			Destination: &config.https.cert,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "key-file",
			Usage:       "Location of the RSA private key `file`",
			Value:       "",
			Destination: &config.https.key,
		}),

		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "enable-auth",
			Usage:       "Enable authentication",
			Value:       false,
			Destination: &config.auth.enabled,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "username",
			Usage:       "Username",
			Value:       "",
			Destination: &config.auth.username,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "password",
			Usage:       "Password",
			Value:       "",
			Destination: &config.auth.password,
		}),

		altsrc.NewIntFlag(&cli.IntFlag{
			Name:        "timeout-read",
			Value:       0,
			Usage:       "Maximum duration for reading the entire request, including the body",
			Destination: &config.timeout.read,
		}),
		altsrc.NewIntFlag(&cli.IntFlag{
			Name:        "timeout-write",
			Value:       0,
			Usage:       "Maximum duration before timing out writes of the response",
			Destination: &config.timeout.write,
		}),
		altsrc.NewIntFlag(&cli.IntFlag{
			Name:        "timeout-dial",
			Value:       10,
			Usage:       "Dial timeout",
			Destination: &config.timeout.dial,
		}),

		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "enable-log",
			Usage:       "Enable file logging",
			Value:       false,
			Destination: &config.log.enabled,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "log-dir",
			Value:       "log",
			Usage:       "Location of the log directory",
			Destination: &config.log.dir,
		}),
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "log-connections",
			Usage:       "Log HTTP(S) connections",
			Value:       true,
			Destination: &config.log.connections,
		}),

		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "quiet",
			Aliases:     []string{"q"},
			Usage:       "Activate quiet mode",
			Value:       false,
			Destination: &config.quiet,
		}),

		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Value:   "",
			Usage:   "Location of the configuration `file` in .yml format",
		},
	}

	cli.HelpFlag = &quietBoolFlag{
		BoolFlag: cli.BoolFlag{
			Name:    "help",
			Aliases: []string{"h"},
			Usage:   "Print this help text and exit",
		},
	}

	cli.VersionFlag = &quietBoolFlag{
		BoolFlag: cli.BoolFlag{
			Name:    "version",
			Aliases: []string{"v"},
			Usage:   "Print program version and exit",
		},
	}

	app.cli = &cli.App{
		Name:                  "HTTP(S) Proxy Server",
		Version:               "v1.0.0",
		Compiled:              time.Now(),
		UsageText:             fmt.Sprintf("%s [global options]", app.name),
		HelpName:              app.name,
		HideHelpCommand:       true,
		Before:                altsrc.InitInputSourceWithContext(flags, altsrc.NewYamlSourceFromFlagFunc("config")),
		Flags:                 flags,
		CustomAppHelpTemplate: helpTemplate,
		Action: func(cCtx *cli.Context) error {
			if len(os.Args) < 2 {
				cli.ShowAppHelp(cCtx)
				os.Exit(0)
			}

			if !cCtx.Bool("help") && !cCtx.Bool("version") {
				if err := config.init(); err != nil {
					return err
				}

				if err := log.init(); err != nil {
					return err
				}

				app.run = true
			}

			return nil
		},
	}
}

type quietBoolFlag struct {
	cli.BoolFlag
}

func (q *quietBoolFlag) String() string {
	return cli.FlagStringer(q)
}

func (q *quietBoolFlag) GetDefaultText() string {
	return ""
}

package main

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/mikesimons/earl"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "pacyak"
	app.Version = "1.0"

	cli.AppHelpTemplate = `{{.Name}} version {{.Version}} - For the unfortunate souls stuck behind corporate proxies

{{.HelpName}} [options] <pac location>

OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
`

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "listen",
			Usage: "Pacyak will listen for requests to this address",
			Value: "127.0.0.1:8080",
		},
		cli.StringFlag{
			Name:  "ping-host",
			Usage: "Host only accessible from within your proxy. Required if PAC location is a file. (default: Host of PAC location)",
		},
		cli.StringFlag{
			Name:  "pac-proxy",
			Usage: "Proxy for pac file. (Only necessary if your PAC location requires a proxy to be set)",
		},
		cli.StringFlag{
			Name:  "log-level",
			Usage: "Log level (debug, info, warn, error)",
			Value: "info",
		},
	}

	app.Action = func(c *cli.Context) error {
		opts := &PacYakOpts{}

		tmp, err := logrus.ParseLevel(c.String("log-level"))
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Invalid log level '%s'. Valid levels are: debug, info, warn, error", tmp), 1)
		}
		opts.LogLevel = tmp

		if c.NArg() < 1 {
			cli.ShowAppHelp(c)
			os.Exit(0)
		}

		opts.PacFile = c.Args().Get(0)

		url := earl.Parse(opts.PacFile)
		if url.Scheme == "" || url.Scheme == "file" {
			if _, err := os.Stat(opts.PacFile); os.IsNotExist(err) {
				return cli.NewExitError("PAC location is not a valid URL and file does not exist. If it is a URL, please specify a protocol (e.g. http://)", 1)
			}
		}

		opts.PingCheckHost = earl.Parse(c.String("ping-host")).Host
		if c.String("ping-host") == "" {
			url = earl.Parse(opts.PacFile)
			if url.Scheme == "" || url.Scheme == "file" {
				return cli.NewExitError("--ping-host is required if PAC location is not a URL", 1)
			}
			opts.PingCheckHost = url.Host
		}

		opts.PacProxy = c.String("pac-proxy")
		opts.ListenAddr = c.String("listen")

		Run(opts)
		return nil
	}

	app.Run(os.Args)
}

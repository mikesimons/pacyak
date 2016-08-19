package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/mikesimons/earl"
)

func main() {
	opts := &PacYakOpts{}

	flag.StringVar(&opts.ListenAddr, "addr", "127.0.0.1:8080", "Proxy listen address")
	flag.StringVar(&opts.ICMPCheckHost, "pinghost", "", "Availability check host")
	flag.StringVar(&opts.PacProxy, "pacproxy", "", "Proxy for fetching PAC file")
	flag.StringVar(&opts.LogLevelStr, "loglevel", "info", "Log level: debug, info, warn, error")
	flag.Parse()

	tmp, err := logrus.ParseLevel(opts.LogLevelStr)
	if err != nil {
		fmt.Printf("Invalid log level '%s'. Valid levels are: debug, info, warn, error")
		os.Exit(1)
	}
	opts.LogLevel = tmp

	args := flag.Args()

	if len(args) < 1 {
		fmt.Printf("Usage: pacyak <pac-file>\n")
		os.Exit(1)
	}

	opts.PacFile = args[0]

	url := earl.Parse(opts.ICMPCheckHost)
	if url.Host == "" {
		url = earl.Parse(opts.PacFile)
		if url.Host == "" {
			fmt.Printf("-pinghost is required if pac-file is not a URL")
			os.Exit(1)
		}
	}

	opts.ICMPCheckHost = url.Host

	app := NewPacYakApp(opts)
	log.Fatal(http.ListenAndServe(opts.ListenAddr, app))
}

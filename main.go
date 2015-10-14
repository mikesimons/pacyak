package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mikesimons/earl"
)

func main() {
	opts := &PacYakOpts{}
	flag.StringVar(&opts.ListenAddr, "addr", "127.0.0.1:8080", "Proxy listen address")
	flag.StringVar(&opts.ICMPCheckHost, "icmphost", "", "Availability check host (ICMP)")
	flag.StringVar(&opts.PacProxy, "pacproxy", "", "Proxy for fetching PAC file")
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		fmt.Printf("Usage: pacyak <pac-file>\n")
		os.Exit(1)
	}

	opts.PacFile = args[0]

	url := earl.Parse(opts.ICMPCheckHost)
	if url.Host == "" {
		fmt.Printf("Invalid ICMP check host: %s", opts.ICMPCheckHost)
		os.Exit(1)
	}

	opts.ICMPCheckHost = url.Host

	app := NewPacYakApp(opts)
	log.Fatal(http.ListenAndServe(opts.ListenAddr, app))
}

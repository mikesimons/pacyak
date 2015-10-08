package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "Proxy listen address")
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <url-to-pac-file>\n", os.Args[0])
		os.Exit(1)
	}
	app := NewPacYakApp(os.Args[1], *addr)
	log.Fatal(http.ListenAndServe(*addr, app))
}

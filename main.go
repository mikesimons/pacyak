package main

import (
  "flag"
  "log"
  "net/http"
  "os"
)

func main() {
  addr := flag.String("addr", "127.0.0.1:8080", "Proxy listen address")
  app := NewPacYakApp(os.Args[1], *addr)
  log.Fatal(http.ListenAndServe(*addr, app))
}

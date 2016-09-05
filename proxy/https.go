package proxy

import (
	"io"
	"net"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

// hijack will attempt to hijack the given response to send raw data
func (proxy *Proxy) hijack(response http.ResponseWriter) (net.Conn, bool) {
	hijacker, ok := response.(http.Hijacker)
	if !ok {
		log.Error("httpserver does not support hijacking")
		response.WriteHeader(502)
		response.Write([]byte("pacyak error: httpserver does not support hijacking!"))
		return nil, false
	}

	hijacked, _, err := hijacker.Hijack()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Could not hijack connection")
		response.WriteHeader(502)
		response.Write([]byte("pacyak error: could not hijack connection!"))
		return nil, false
	}

	return hijacked, true
}

// connectDial connects to the given addr for a CONNECT request using either an overridden dialer or the default if not set.
// Derived from github.com/elazarl/go-proxy
func (proxy *Proxy) connectDial(network, addr string) (c net.Conn, err error) {
	if proxy.ConnectDial == nil {
		return proxy.Tr.Dial(network, addr)
	}
	return proxy.ConnectDial(network, addr)
}

// copyAndClose pumps data from one connection to the other and closes once data ceases flowing.
// Derived from github.com/elazarl/go-proxy
func copyAndClose(w, r net.Conn) {
	// Lots of "read connection reset by peer" errs if we both with the error here
	// That's because the server may terminate connection at will
	// There is nothing we can do about that so we ignore it
	io.Copy(w, r)
	err := r.Close()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Error closing connection")
	}
}

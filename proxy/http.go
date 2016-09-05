package proxy

import (
	"io"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

// filterRequestHeaders removes headers that a proxy should remove
// Derived from github.com/elazarl/go-proxy
func (proxy *Proxy) filterRequestHeaders(r *http.Request) {
	r.Header.Del("Accept-Encoding")
	r.Header.Del("Proxy-Connection")
	r.Header.Del("Proxy-Authenticate")
	r.Header.Del("Proxy-Authorization")
	r.Header.Del("Connection")
}

// makeUpstreamRequest roundtrips the given request and handles errors from that
func (proxy *Proxy) makeUpstreamRequest(request *http.Request) (*http.Response, bool) {
	response, err := proxy.Tr.RoundTrip(request)

	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Error performing roundtrip")
	}

	return response, err == nil
}

// copyResponse copies headers, status and body from an upstream request to a response
// Derived from github.com/elazarl/go-proxy
func (proxy *Proxy) copyResponse(upstream *http.Response, response http.ResponseWriter) {
	// Copy headers
	headers := response.Header()
	for header, values := range upstream.Header {
		headers.Del(header)
		for _, value := range values {
			headers.Add(header, value)
		}
	}

	// Copy status code & body
	response.WriteHeader(upstream.StatusCode)
	_, err := io.Copy(response, upstream.Body)

	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Error copying body")
	}

	err = upstream.Body.Close()

	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Error closing body")
	}
}

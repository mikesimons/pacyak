package proxy

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mikesimons/earl"
)

// Proxy is a simple proxy implementation
type Proxy struct {
	Tr            *http.Transport
	DirectHandler http.Handler
	ConnectDial   func(network string, addr string) (net.Conn, error)
	Logger        *log.Logger
	Available     func() bool
}

// connectDialer establishes a connection for use with a CONNECT request
// This code is largely derived from github.com/elazarl/go-proxy
func (proxy *Proxy) connectDialer(https_proxy string) func(network, addr string) (net.Conn, error) {
	u := earl.ParseWithDefaults(https_proxy, &earl.URL{Scheme: "auto", Port: "80"})

	return func(network, addr string) (net.Conn, error) {
		client, err := proxy.Tr.Dial(network, u.HostAndPort())
		if err != nil {
			return nil, fmt.Errorf("Proxy refused connection: %s", err)
		}

		if u.Scheme == "https" {
			client = tls.Client(client, proxy.Tr.TLSClientConfig)
		}

		request := &http.Request{
			Method: "CONNECT",
			URL:    &url.URL{Opaque: addr},
			Host:   addr,
			Header: make(http.Header),
		}

		request.Write(client)

		reader := bufio.NewReader(client)
		response, err := http.ReadResponse(reader, request)
		if err != nil {
			client.Close()
			return nil, fmt.Errorf("Error reading response from proxy: %s", err)
		}

		if response.StatusCode != 200 {
			responseText, _ := ioutil.ReadAll(response.Body)
			response.Body.Close()
			client.Close()
			return nil, fmt.Errorf("Proxy error: %s", string(responseText))
		}

		return client, nil
	}
}

// New creates a new instance of Proxy. "direct" is a special case URL that simply passes data through.
func New(proxyURLString string) *Proxy {
	proxy := &Proxy{
		Tr: &http.Transport{
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			IdleConnTimeout: 90 * time.Second,
			Dial:            net.Dial,
		},
	}

	if proxyURLString == "direct" {
		proxy.Tr.Proxy = func(req *http.Request) (*url.URL, error) { return nil, nil }
		proxy.Available = func() bool { return true }
		proxy.ConnectDial = nil
	} else {
		proxyURL := earl.ParseWithDefaults(proxyURLString, &earl.URL{Scheme: "auto"})
		proxy.Tr.Proxy = func(req *http.Request) (*url.URL, error) { return proxyURL.ToNetURL(), nil }
		proxy.Available = func() bool { return exec.Command("ping", "-w", "1", proxyURL.Host).Run() == nil }
		proxy.ConnectDial = proxy.connectDialer(proxyURL.ToNetURL().String())
	}

	return proxy
}

// ServeHTTP handles the actual http / https proxying
// Derived from github.com/elazarl/go-proxy
func (proxy *Proxy) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if request.Method == "CONNECT" {
		remote, err := proxy.connectDial("tcp", request.URL.Host)
		if err != nil {
			log.WithFields(log.Fields{"host": request.URL.Host, "error": err}).Error("Unable to connect to remote host")
			response.WriteHeader(502)

			// TODO nicer error page
			response.Write([]byte("pacyak error: unable to connect to remote host"))
			return
		}

		hijacked, ok := proxy.hijack(response)
		if !ok {
			return
		}

		hijacked.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))

		go copyAndClose(remote, hijacked)
		go copyAndClose(hijacked, remote)
	} else {
		proxy.filterRequestHeaders(request)
		if upstreamResponse, ok := proxy.makeUpstreamRequest(request); ok {
			proxy.copyResponse(upstreamResponse, response)
		}
	}
}

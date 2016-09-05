package proxyfactory

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mikesimons/goproxy"
	"github.com/mikesimons/earl"
)

// ProxyFactory holds all state for the proxy factory
type ProxyFactory struct {
	proxies            map[string]*goproxy.ProxyHttpServer
	availabilityChecks map[string](func() bool)
	availability       map[string]bool
	Logger             *log.Logger
}

// New is the constructor function for ProxyFactory
func New() *ProxyFactory {
	pf := &ProxyFactory{
		proxies:            make(map[string]*goproxy.ProxyHttpServer),
		availabilityChecks: make(map[string](func() bool)),
		availability:       make(map[string]bool),
		Logger:             log.New(),
	}

	go func() {
		for _ = range time.Tick(30 * time.Second) {
			for key, check := range pf.availabilityChecks {
				pf.availability[key] = check()
				pf.Logger.WithFields(log.Fields{
					"proxy":     key,
					"available": pf.availability[key],
				}).Info("Proxy availability check")
			}
		}
	}()

	return pf
}

func (pf *ProxyFactory) available(handle string) bool {
	if _, ok := pf.availability[handle]; !ok {
		return false
	}

	return pf.availability[handle]
}

// Proxy will return an instance of a goproxy based on the handle
// If one already exists with the given handle, it will be used.
// Otherwise a new one will be created.
func (pf *ProxyFactory) Proxy(handle string) *goproxy.ProxyHttpServer {
	if _, ok := pf.proxies[handle]; !ok {
		proxyURL := earl.Parse(handle)
		if proxyURL.Scheme == "" {
			proxyURL.Scheme = "http"
		}

		proxy := goproxy.NewProxyHttpServer()
		proxy.Tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy: func(req *http.Request) (*url.URL, error) {
				if handle == "direct" {
					return nil, nil
				}

				return proxyURL.ToNetURL(), nil
			},
		}

		if handle == "direct" {
			proxy.ConnectDial = nil
		} else {
			proxy.ConnectDial = proxy.NewConnectDialToProxy(proxyURL.ToNetURL().String())
		}

		proxy.Verbose = true

		pf.availabilityChecks[handle] = func() bool {
			if handle == "direct" {
				return true
			}

			return exec.Command("ping", "-w", "1", proxyURL.Host).Run() == nil
		}

		//proxy.Logger = pf.Logger

		pf.availability[handle] = pf.availabilityChecks[handle]()
		pf.proxies[handle] = proxy
	}

	return pf.proxies[handle]
}

// FromPacResponse takes a PAC response string and returns a goproxy
func (pf *ProxyFactory) FromPacResponse(response string) *goproxy.ProxyHttpServer {
	if response == "DIRECT" {
		return pf.Proxy("direct")
	}

	response = strings.Replace(response, "PROXY", "", -1)
	response = strings.Replace(response, " ", "", -1)
	proxies := strings.Split(response, ";")
	for _, proxyStr := range proxies {
		proxy := pf.Proxy(proxyStr)

		if !pf.available(proxyStr) {
			continue
		}

		return proxy
	}

	return pf.Proxy("direct")
}

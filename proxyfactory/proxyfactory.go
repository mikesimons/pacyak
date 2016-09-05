package proxyfactory

import (
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mikesimons/pacyak/proxy"
)

// ProxyFactory holds all state for the proxy factory
type ProxyFactory struct {
	proxies      map[string]*proxy.Proxy
	availability map[string]bool
}

// New is the constructor function for ProxyFactory
func New() *ProxyFactory {
	pf := &ProxyFactory{
		proxies:      make(map[string]*proxy.Proxy),
		availability: make(map[string]bool),
	}

	go func() {
		for _ = range time.Tick(30 * time.Second) {
			for key, proxy := range pf.proxies {
				pf.availability[key] = proxy.Available()
				log.WithFields(log.Fields{
					"proxy":     key,
					"available": pf.availability[key],
				}).Debug("Proxy availability check")
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

// Proxy will return an instance of a proxy based on the handle
// If one already exists with the given handle, it will be used.
// Otherwise a new one will be created.
func (pf *ProxyFactory) Proxy(handle string) *proxy.Proxy {
	if _, ok := pf.proxies[handle]; !ok {
		proxy := proxy.New(handle)
		pf.availability[handle] = proxy.Available()
		pf.proxies[handle] = proxy
	}

	return pf.proxies[handle]
}

// FromPacResponse takes a PAC response string and returns a proxy
func (pf *ProxyFactory) FromPacResponse(response string) *proxy.Proxy {
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

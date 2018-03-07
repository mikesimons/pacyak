package proxyfactory

import (
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"../../pacyak/proxy"
)

// ProxyFactory holds all state for the proxy factory
type ProxyFactory struct {
	proxies      map[string]*proxy.Proxy
	availability map[string]bool
	lock         *sync.Mutex
}

// New is the constructor function for ProxyFactory
func New() *ProxyFactory {
	pf := &ProxyFactory{
		proxies:      make(map[string]*proxy.Proxy),
		availability: make(map[string]bool),
		lock:         &sync.Mutex{},
	}

	go func() {
		for _ = range time.Tick(30 * time.Minute) {
			pf.lock.Lock()
			for key, proxy2 := range pf.proxies {
				pf.availability[key] = proxy2.Available()

				log.WithFields(log.Fields{
					"proxy":     key,
					"available": pf.availability[key],
				}).Debug("Proxy availability check")
			}
			pf.lock.Unlock()
		}
	}()

	return pf
}

func (pf *ProxyFactory) available(handle string) bool {
	pf.lock.Lock()
	defer func() {
		pf.lock.Unlock()
	}()

	if _, ok := pf.availability[handle]; !ok {
		return false
	}

	return pf.availability[handle]
}

// Proxy will return an instance of a proxy based on the handle
// If one already exists with the given handle, it will be used.
// Otherwise a new one will be created.
func (pf *ProxyFactory) Proxy(handle string) *proxy.Proxy {
	var ret *proxy.Proxy

	pf.lock.Lock()
	if _, ok := pf.proxies[handle]; !ok {
		log.WithFields(log.Fields{
			"proxy":     handle,
		}).Debug("Creating new proxy for handle")
		newProxy := proxy.New(handle)
		pf.availability[handle] = newProxy.Available()
		pf.proxies[handle] = newProxy
	}
	ret = pf.proxies[handle]
	pf.lock.Unlock()

	return ret
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
		proxy2 := pf.Proxy(proxyStr)

		if !pf.available(proxyStr) {
			continue
		}

		return proxy2
	}

	return pf.Proxy("direct")
}

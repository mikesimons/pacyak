package proxyfactory

import (
  "github.com/elazarl/goproxy"
  log "github.com/Sirupsen/logrus"
  "github.com/mikesimons/earl"
  "net/url"
  "net/http"
  "crypto/tls"
  "strings"
  "os/exec"
  "time"
)

type ProxyFactory struct {
  proxies map[string]*goproxy.ProxyHttpServer
  availabilityChecks map[string](func() bool)
  availability map[string]bool
  Logger *log.Logger
}

func New() (*ProxyFactory) {
  pf := &ProxyFactory{
    proxies: make(map[string]*goproxy.ProxyHttpServer),
    availabilityChecks: make(map[string](func() bool)),
    availability: make(map[string]bool),
    Logger: log.New(),
  }

  go func() {
    for _ = range time.Tick(30 * time.Second) {
      for key, check := range pf.availabilityChecks {
        pf.availability[key] = check()
        pf.Logger.WithFields(log.Fields{
          "proxy": key,
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

func (pf *ProxyFactory) Proxy(handle string) (*goproxy.ProxyHttpServer) {
  if _, ok := pf.proxies[handle]; !ok {
    proxyUrl := earl.Parse(handle)
    if proxyUrl.Scheme == "" {
      proxyUrl.Scheme = "http"
    }

    proxy := goproxy.NewProxyHttpServer()
    proxy.Tr = &http.Transport{
  		TLSClientConfig: &tls.Config{ InsecureSkipVerify: true },
  		Proxy: func(req *http.Request) (*url.URL, error) {
        if handle == "direct" {
          return nil, nil
        } else {
  			  return proxyUrl.ToNetUrl(), nil
        }
  		},
  	}

    if handle == "direct" {
      proxy.ConnectDial = nil
    } else {
      proxy.ConnectDial = proxy.NewConnectDialToProxy(proxyUrl.ToNetUrl().String())
    }

    pf.availabilityChecks[handle] = func() bool {
      if handle == "direct" {
        return true
      } else {
        return exec.Command("ping", "-w", "1", proxyUrl.Host).Run() == nil
      }
    }

    //proxy.Logger = pf.Logger

    pf.availability[handle] = pf.availabilityChecks[handle]()
    pf.proxies[handle] = proxy
  }

  return pf.proxies[handle]
}

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

    pf.Logger.WithFields(log.Fields{ "proxy": proxyStr }).Info("Using proxy")

    return proxy
  }

  return pf.Proxy("direct")
}

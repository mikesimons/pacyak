package main

import (
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mikesimons/earl"
	"github.com/mikesimons/pacyak/pacsandbox"
	"github.com/mikesimons/pacyak/proxyfactory"
	"github.com/mikesimons/readly"
)

const DIRECT_SANDBOX = 0
const PROXY_SANDBOX = 1

// PacYakOpts holds runtime config options for PacYakApplication
type PacYakOpts struct {
	PingCheckHost string
	PacFile       string
	ListenAddr    string
	PacProxy      string
	LogLevelStr   string
	LogLevel      log.Level
}

// pacInterpreter is a simple interface we use to provide a dummy implementation of pacsandbox for directPac
type pacInterpreter interface {
	ProxyFor(string) (string, error)
	Reset() // HACK
}

// directPac is a dummy implementation of pacsandbox to avoid invoking JS to return a constant static string ("DIRECT")
type directPac struct{}

func (p *directPac) ProxyFor(s string) (string, error) { return "DIRECT", nil }
func (p *directPac) Reset()                            {}

// PacYakApplication holds all application state
type PacYakApplication struct {
	opts         *PacYakOpts
	pacFile      *earl.URL
	sandboxIndex int
	sandboxes    []pacInterpreter
	factory      *proxyfactory.ProxyFactory
	listenAddr   string
	interfaceMap map[string]string
	Reader       *readly.Reader
}

// Run is the entry point for pacyak. It will initialize pacyak and start listening.
func Run(opts *PacYakOpts) {

	log.SetLevel(opts.LogLevel)
	reader := readly.New()

	// We need to explicitly set HTTP client to prevent it trying to use ENV vars for proxy
	// pacyak listen addr is expected to be set as HTTP_PROXY / HTTPS_PROXY but it isn't started yet!
	// This level of control also means a lib like hashicorp/go-getter is not suitable :(
	reader.Client = &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				if opts.PacProxy != "" {
					return url.Parse(opts.PacProxy)
				}

				return nil, nil
			},
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 5 * time.Second,
			}).DialContext,
			IdleConnTimeout: 5 * time.Second,
		},
	}

	app := &PacYakApplication{
		opts:         opts,
		pacFile:      earl.Parse(opts.PacFile),
		factory:      proxyfactory.New(),
		sandboxIndex: DIRECT_SANDBOX,
		sandboxes:    []pacInterpreter{&directPac{}, &directPac{}},
		listenAddr:   opts.ListenAddr,
		Reader:       reader,
	}

	go app.monitorPingAvailability()
	go app.monitorNetworkInterfaces()

	// FIXME - graceful handler; server 502 on error and keep going
	log.Fatal(http.ListenAndServe(app.opts.ListenAddr, app))
}

// ServeHTTP handles directing the request to the correct proxy
func (app *PacYakApplication) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{
		"method": r.Method,
		"url":    r.URL.String(),
	}).Debug("Processing HTTP request")

	// FIXME: We want to be able to serve some stats / tools from here
	//if !r.URL.IsAbs() {
	//	fmt.Fprintf(w, `function FindProxyForURL(url, host) { return "PROXY %s"; }`, app.listenAddr)
	//	return
	//}

	pacResponse, err := app.activeSandbox().ProxyFor(r.URL.String())

	if err != nil {
		log.WithFields(log.Fields{"response": pacResponse, "sandbox_error": err, "url": r.URL.String()}).Error("Sandbox error!")
	} else {
		log.WithFields(log.Fields{"response": pacResponse}).Debug("PAC result")
	}

	proxy := app.factory.FromPacResponse(pacResponse)

	proxy.ServeHTTP(w, r)
}

// switchToDirect switches the pac sandbox to the dummy "DIRECT" implementation
// We do this when our ping check fails (indicating a proxy may no longer be required)
func (app *PacYakApplication) switchToDirect() {
	if app.sandboxIndex != DIRECT_SANDBOX {
		log.Info("PAC availability check failed; switching to direct")
		app.setSandbox(DIRECT_SANDBOX)
	}
}

// switchToPac switches the pac sandbox to the JS implementation (using the PAC file specified on the CLI)
// We do this when our ping check passes (indicating we're in an env that requires a proxy)
func (app *PacYakApplication) switchToPac() {
	if app.sandboxIndex == DIRECT_SANDBOX {
		pac, err := app.Reader.Read(app.pacFile.Input)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("PAC availability check passed but was unable to fetch PAC")
		} else {
			log.Info("PAC availability check passed; switching from direct")
			app.sandboxes[PROXY_SANDBOX] = pacsandbox.New(pac)
			app.setSandbox(PROXY_SANDBOX)
		}
	}
}

// handlePacAvailability updates the status of the ping check and switches sandbox if required
// This may be called from multiple go routines so we wrap it in a mutex to avoid racing
// Tried this with channels once but CPU usage blew up! Probably PEBKAC
func (app *PacYakApplication) handlePacAvailability() {
	available := exec.Command("ping", "-w", "1", app.opts.PingCheckHost).Run() == nil
	log.WithFields(log.Fields{"available": available}).Info("PAC availability check")

	if !available {
		app.switchToDirect()
	} else {
		app.switchToPac()
	}
}

// monitorPingAvailability is a wrapper for handlePacAvailability invoking it every 30 seconds
func (app *PacYakApplication) monitorPingAvailability() {
	app.handlePacAvailability()
	for _ = range time.Tick(30 * time.Second) {
		app.handlePacAvailability()
	}
}

// checkNetworkInterfaces will trigger a ping check if network interfaces have changed since last check
func (app *PacYakApplication) checkNetworkInterfaces() {
	interfaceMap := makeInterfaceMap()
	lastInterfaceMap := app.interfaceMap

	defer func() {
		app.interfaceMap = interfaceMap
	}()

	newInterfaces := interfaceMapKeys(interfaceMap)
	oldInterfaces := interfaceMapKeys(lastInterfaceMap)

	if interfaceListChanged(newInterfaces, oldInterfaces) {
		log.WithFields(log.Fields{"old": oldInterfaces, "new": newInterfaces}).Debug("Network interface list has changed")
		app.handlePacAvailability()
		return
	}

	for key, val := range interfaceMap {
		if lastInterfaceMap[key] != val {
			log.WithFields(log.Fields{"interface": key, "old": lastInterfaceMap[key], "new": val}).Debug("Network interface configuration has changed")
			app.handlePacAvailability()
			return
		}
	}

	log.Debug("No network changes detected")
}

// monitorNetworkInterfaces is a wrapper for checkNetworkInterfaces invoking it every 5 seconds
func (app *PacYakApplication) monitorNetworkInterfaces() {
	app.interfaceMap = makeInterfaceMap()
	for _ = range time.Tick(5 * time.Second) {
		app.checkNetworkInterfaces()
	}
}

func (app *PacYakApplication) activeSandbox() pacInterpreter {
	return app.sandboxes[app.sandboxIndex]
}

func (app *PacYakApplication) setSandbox(index int) {
	app.sandboxes[index].Reset()
	app.sandboxIndex = index
}

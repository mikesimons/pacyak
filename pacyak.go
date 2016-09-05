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

// PacYakOpts holds runtime config options for PacYakApplication
type PacYakOpts struct {
	PingCheckHost string
	PacFile       string
	ListenAddr    string
	PacProxy      string
	LogLevelStr   string
	LogLevel      log.Level
}

type pacInterpreter interface {
	ProxyFor(string) (string, error)
	Reset() // HACK
}

type directPac struct{}

func (p *directPac) ProxyFor(s string) (string, error) { return "DIRECT", nil }
func (p *directPac) Reset()                            {}

// PacYakApplication holds all application state
type PacYakApplication struct {
	opts          *PacYakOpts
	pacFile       *earl.URL
	directSandbox pacInterpreter
	factory       *proxyfactory.ProxyFactory
	sandbox       pacInterpreter
	listenAddr    string
	Reader        *readly.Reader
}

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
		opts:          opts,
		pacFile:       earl.Parse(opts.PacFile),
		factory:       proxyfactory.New(),
		sandbox:       &directPac{},
		directSandbox: &directPac{},
		listenAddr:    opts.ListenAddr,
		Reader:        reader,
	}

	app.startAvailabilityChecks()

	// FIXME - graceful handler; server 502 on error and keep going
	log.Fatal(http.ListenAndServe(app.opts.ListenAddr, app))
}

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

	pacResponse, err := app.sandbox.ProxyFor(r.URL.String())

	if err != nil {
		log.WithFields(log.Fields{"response": pacResponse, "sandbox_error": err, "url": r.URL.String()}).Error("Sandbox error!")
	} else {
		log.WithFields(log.Fields{"response": pacResponse}).Debug("PAC result")
	}

	proxy := app.factory.FromPacResponse(pacResponse)

	proxy.ServeHTTP(w, r)
}

func (app *PacYakApplication) switchToDirect() {
	if app.sandbox != app.directSandbox {
		log.Info("PAC availability check failed; switching to direct")
		app.sandbox = app.directSandbox
		app.sandbox.Reset()
	}
}

func (app *PacYakApplication) switchToPac() {
	if app.sandbox == app.directSandbox {
		pac, err := app.Reader.Read(app.pacFile.Input)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("PAC availability check passed but was unable to fetch PAC")
		} else {
			log.Info("PAC availability check passed; switching from direct")
			sandbox := pacsandbox.New(pac)
			app.sandbox = sandbox
		}
	}
}

func (app *PacYakApplication) handlePacAvailability() {
	available := exec.Command("ping", "-w", "1", app.opts.PingCheckHost).Run() == nil
	log.WithFields(log.Fields{"available": available}).Info("PAC availability check")

	if !available {
		app.switchToDirect()
	} else {
		app.switchToPac()
	}
}

func (app *PacYakApplication) startAvailabilityChecks() {
	app.handlePacAvailability()
	go func() {
		for _ = range time.Tick(30 * time.Second) {
			app.handlePacAvailability()
		}
	}()
}

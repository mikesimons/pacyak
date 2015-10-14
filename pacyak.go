package main

import (
	"fmt"
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
	ICMPCheckHost string
	PacFile       string
	ListenAddr    string
	PacProxy      string
}

// PacYakApplication holds all application state
type PacYakApplication struct {
	opts          *PacYakOpts
	pacFile       *earl.URL
	directSandbox *pacsandbox.PacSandbox
	factory       *proxyfactory.ProxyFactory
	sandbox       *pacsandbox.PacSandbox
	listenAddr    string
	Logger        *log.Logger
	Reader        *readly.Reader
}

func (app *PacYakApplication) switchToDirect() {
	if app.sandbox != app.directSandbox {
		app.Logger.Info("PAC availability check failed; switching to direct")
		app.sandbox = app.directSandbox
		app.sandbox.PurgeCache()
	}
}

func (app *PacYakApplication) switchToPac() {
	if app.sandbox == app.directSandbox {
		pac, err := app.Reader.Read(app.pacFile.Input)
		if err != nil {
			app.Logger.WithFields(log.Fields{"error": err}).Error("PAC availability check passed but was unable to fetch PAC")
		} else {
			app.Logger.Info("PAC availability check passed; switching from direct")
			app.sandbox = pacsandbox.New(pac)
			app.sandbox.Logger = app.Logger
		}
	}
}

func (app *PacYakApplication) handlePacAvailability() {
	available := exec.Command("ping", "-w", "1", app.opts.ICMPCheckHost).Run() == nil
	app.Logger.WithFields(log.Fields{"available": available}).Info("PAC availability check")

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

// NewPacYakApp will create a new PacYakApplication instance
func NewPacYakApp(opts *PacYakOpts) *PacYakApplication {
	logger := log.New()

	directSandbox := pacsandbox.New(`function FindProxyForURL(url, host) { return "DIRECT"; }`)
	directSandbox.Logger = logger

	// We need to explicitly set HTTP client to prevent it trying to use ENV vars for proxy
	// pacyak listen addr is expected to be set as HTTP_PROXY / HTTPS_PROXY but it isn't started yet!
	reader := readly.New()
	reader.Client = &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				if opts.PacProxy != "" {
					return url.Parse(opts.PacProxy)
				}

				return nil, nil
			},
		},
	}

	application := &PacYakApplication{
		opts:          opts,
		pacFile:       earl.Parse(opts.PacFile),
		factory:       proxyfactory.New(),
		sandbox:       directSandbox,
		directSandbox: directSandbox,
		listenAddr:    opts.ListenAddr,
		Logger:        logger,
		Reader:        reader,
	}

	log.SetLevel(log.InfoLevel)

	application.startAvailabilityChecks()

	return application
}

func (app *PacYakApplication) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.Logger.WithFields(log.Fields{
		"method": r.Method,
		"url":    r.URL.String(),
	}).Info("Processing HTTP request")

	if r.URL.Path == "/pac" {
		fmt.Fprintf(w, `function FindProxyForURL(url, host) { return "PROXY %s"; }`, app.listenAddr)
		return
	}

	pacResponse, err := app.sandbox.ProxyFor(r.URL.String())

	if err != nil {
		app.Logger.WithFields(log.Fields{"response": pacResponse, "sandbox_error": err, "url": r.URL.String()}).Error("Sandbox error!")
	} else {
		app.Logger.WithFields(log.Fields{"response": pacResponse}).Info("PAC result")
	}

	proxy := app.factory.FromPacResponse(pacResponse)

	proxy.ServeHTTP(w, r)
}

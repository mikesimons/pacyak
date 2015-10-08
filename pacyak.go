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

// PacYakApplication holds all application state
type PacYakApplication struct {
	pacURL        *earl.URL
	directSandbox *pacsandbox.PacSandbox
	factory       *proxyfactory.ProxyFactory
	sandbox       *pacsandbox.PacSandbox
	listenAddr    string
	Logger        *log.Logger
}

func loadPac(file string) (string, error) {
	readly.SetClient(&http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return nil, nil
			},
		},
	})

	r, err := readly.Read(file)

	if err != nil {
		return "", err
	}

	return r, nil
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
		pac, err := loadPac(app.pacURL.ToNetURL().String())
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
	available := exec.Command("ping", "-w", "1", app.pacURL.Host).Run() == nil
	app.Logger.WithFields(log.Fields{"available": available}).Info("PAC URL check")

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
func NewPacYakApp(pacURLStr string, listenAddr string) *PacYakApplication {
	// TODO Support local PAC files w/ external check URL
	logger := log.New()

	directSandbox := pacsandbox.New(`function FindProxyForURL(url, host) { return "DIRECT"; }`)
	directSandbox.Logger = logger

	application := &PacYakApplication{
		pacURL:        earl.Parse(pacURLStr),
		factory:       proxyfactory.New(),
		sandbox:       directSandbox,
		directSandbox: directSandbox,
		listenAddr:    listenAddr,
		Logger:        logger,
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

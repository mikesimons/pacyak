package pacsandbox

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mikesimons/earl"
	"github.com/robertkrimen/otto"
	"github.com/wunderlist/ttlcache"
)

// PacSandbox holds state for the pac sandbox instance
type PacSandbox struct {
	pac         string
	vm          *otto.Otto
	cache       *ttlcache.Cache // TODO rename
	resultCache *ttlcache.Cache
	Logger      *log.Logger
}

// New is the constructor for PacSandbox
func New(pac string) *PacSandbox {
	sandbox := &PacSandbox{
		pac:    pac,
		vm:     otto.New(),
		Logger: log.New(),
	}

	sandbox.Reset()
	sandbox.initPacFunctions()
	sandbox.vm.Run(pac)

	return sandbox
}

// ProxyFor will take a URL, run it through the PAC logic and produce a PAC result string
func (p *PacSandbox) ProxyFor(u string) (string, error) {
	parsedURL := earl.Parse(u)

	key := fmt.Sprintf("%s-%s-%s-result", parsedURL.Scheme, parsedURL.Host, parsedURL.Port)
	if val, ok := p.resultCache.Get(key); ok {
		p.Logger.WithFields(log.Fields{"key": key}).Info("PacSandbox result cache hit")
		return val, nil
	}

	js := fmt.Sprintf(
		"FindProxyForURL(%#v, %#v);",
		u,
		parsedURL.Host,
	)

	vm := p.vm.Copy()
	result, err := p.ottoRetString(
		vm.Run(js),
	)

	if err == nil {
		p.resultCache.Set(key, result)
	}

	return result, err
}

// PurgeCache will (re)initialize internal caches
func (p *PacSandbox) Reset() {
	p.cache = ttlcache.NewCache(5 * time.Minute)
	p.resultCache = ttlcache.NewCache(30 * time.Second)
}

package pacsandbox

import(
  log "github.com/Sirupsen/logrus"
  "github.com/robertkrimen/otto"
  "github.com/wunderlist/ttlcache"
  "github.com/mikesimons/earl"
  "fmt"
  "time"
)

type PacSandbox struct {
  pac string
  vm *otto.Otto
  cache *ttlcache.Cache // TODO rename
  resultCache *ttlcache.Cache
  Logger *log.Logger
}

func New(pac string) *PacSandbox {
  sandbox := &PacSandbox{
    pac: pac,
    vm: otto.New(),
    Logger: log.New(),
  }

  sandbox.PurgeCache()
  sandbox.initPacFunctions()
  sandbox.vm.Run(pac)

  return sandbox
}

func (p *PacSandbox) ProxyFor(u string) (string, error) {
  parsedUrl := earl.Parse(u)

  key := fmt.Sprintf("%s-%s-%s-result", parsedUrl.Scheme, parsedUrl.Host, parsedUrl.Port)
  if val, ok := p.resultCache.Get(key); ok {
    p.Logger.WithFields(log.Fields{ "key": key }).Info("PacSandbox result cache hit")
    return val, nil
  }

  js := fmt.Sprintf(
    "FindProxyForURL(%#v, %#v);",
    u,
    parsedUrl.Host,
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

func (p *PacSandbox) PurgeCache() {
  p.cache = ttlcache.NewCache(5 * time.Minute)
  p.resultCache = ttlcache.NewCache(30 * time.Second)
}

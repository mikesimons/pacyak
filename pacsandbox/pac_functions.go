package pacsandbox

// Reference: http://lxr.mozilla.org/seamonkey/source/netwerk/base/src/nsProxyAutoConfig.js#189

import(
  "github.com/robertkrimen/otto"
  "net"
  "strings"
  "regexp"
  "fmt"
)

func (p *PacSandbox) initPacFunctions() {
  p.vm.Set("dnsResolve", func(call otto.FunctionCall) otto.Value {
    args := p.ottoStringArgs(call, 1, "dnsResolve")
    rval, err := p.dnsResolve(args[0])
    if rval == "" {
      return p.ottoRetValue(false, err)
    } else {
      return p.ottoRetValue(rval, err)
    }
  })

  p.vm.Set("dnsDomainIs", func(call otto.FunctionCall) otto.Value {
    args := p.ottoStringArgs(call, 2, "dnsDomainIs")
    return p.ottoRetValue(
      p.dnsDomainIs(args[0], args[1]),
    )
  })

  p.vm.Set("isResolvable", func(call otto.FunctionCall) otto.Value {
    args := p.ottoStringArgs(call, 1, "isResolvable")
    return p.ottoRetValue(
      p.isResolvable(args[0]),
    )
  })

  p.vm.Set("shExpMatch", func(call otto.FunctionCall) otto.Value {
    args := p.ottoStringArgs(call, 2, "shExpMatch")
    return p.ottoRetValue(
      p.shExpMatch(args[0], args[1]),
    )
  })

  p.vm.Set("isInNet", func(call otto.FunctionCall) otto.Value {
    args := p.ottoStringArgs(call, 3, "isInNet")
    return p.ottoRetValue(
      p.isInNet(args[0], args[1], args[2]),
    )
  })

  p.vm.Set("isPlainHostName", func(call otto.FunctionCall) otto.Value {
    args := p.ottoStringArgs(call, 1, "isPlainHostName")
    return p.ottoRetValue(
      p.isPlainHostName(args[0]),
    )
  })
}

func (p *PacSandbox) dnsDomainIs(host string, domain string) (bool, error) {
  return strings.HasSuffix(host, domain), nil
}

func (p *PacSandbox) dnsResolve(host string) (string, error) {
  if cached, ok := p.cache.Get(host); ok {
    return cached, nil
  }

  if net.ParseIP(host) != nil {
    return host, nil
  }

  result, err := net.LookupHost(host)

  if err != nil {
    return "", nil
  }

  p.cache.Set(host, result[0])
  return result[0], nil
}

func (p *PacSandbox) isResolvable(host string) (bool, error) {
  r, err := p.dnsResolve(host)
  return err == nil && r != "", nil
}

func (p *PacSandbox) shExpMatch(str string, pattern string) (bool, error) {
  pattern = regexp.QuoteMeta(pattern)
  pattern = strings.Replace(pattern, `\?`, ".", -1)
  pattern = strings.Replace(pattern, `\*`, ".*", -1)
  pattern = fmt.Sprintf("^%s$", pattern)
  r := regexp.MustCompile(pattern)

  return r.MatchString(str), nil
}

func (p *PacSandbox) isInNet(ipStr string, ipRangeStr string, ipMaskStr string) (bool, error) {
  ipNet := &net.IPNet{
    IP: net.ParseIP(ipRangeStr),
    Mask: net.IPMask(net.ParseIP(ipMaskStr)),
  }

  return ipNet.Contains(net.ParseIP(ipStr)), nil
}

func (p *PacSandbox) isPlainHostName(host string) (bool, error) {
  return strings.Count(host, ".") == 0, nil
}

package pacsandbox_test

import (
	. "../../pacyak/pacsandbox"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PacSandbox", func() {
	Describe("New", func() {
		It("should create a new instance of PacSandbox", func() {
			it := New("")
			Expect(it).Should(BeAssignableToTypeOf(&PacSandbox{}))
		})

		// should do what if FindProxyForURL not defined?
		// should do what on syntax error?
	})

	Describe("ProxyFor", func() {
		It("should return a proxy for a url", func() {
			it := New(`function FindProxyForURL(url, host) { return "DIRECT"; }`)
			Expect(it.ProxyFor("http://google.com")).Should(Equal("DIRECT"))
		})

		It("should use PAC to return correct proxy", func() {
			it := New(`
			function FindProxyForURL(url, host) {
				if(host == "google.com") {
					return "DIRECT";
				} else {
					return "PROXY hp.com";
				}
			}`)
			Expect(it.ProxyFor("http://google.com")).Should(Equal("DIRECT"))
			Expect(it.ProxyFor("http://hp.com")).Should(Equal("PROXY hp.com"))
		})

		// should do what for invalid url?

		Describe("dnsResolve", func() {
			It("should resolve a hostname to an IP", func() {
				it := New(`function FindProxyForURL(url, host) { return dnsResolve(host); }`)
				Expect(it.ProxyFor("http://google-public-dns-a.google.com")).Should(Equal("8.8.8.8"))
			})

			It("should panic for invalid hostname", func() {
				it := New(`function FindProxyForURL(url, host) { return dnsResolve(host); }`)
				Expect(func() { it.ProxyFor("http://blah.blah.gobble") }).Should(Panic())
			})

			// should do what if not passed anything?
		})

		Describe("dnsDomainIs", func() {
			It("should return boolean indicating if host is subdomain", func() {
				it := New(`function FindProxyForURL(url, host) { return dnsDomainIs(host, "google.com"); }`)
				Expect(it.ProxyFor("http://google.com")).Should(Equal("true"))
				Expect(it.ProxyFor("http://hp.com")).Should(Equal("false"))
			})

			It("should panic with anything less than 2 args", func() {
				it := New(`function FindProxyForURL(url, host) { return dnsDomainIs(); }`)
				Expect(func() { it.ProxyFor("http://google.com") }).Should(Panic())

				it = New(`function FindProxyForURL(url, host) { return dnsDomainIs(1); }`)
				Expect(func() { it.ProxyFor("http://google.com") }).Should(Panic())
			})
		})

		Describe("isResolvable", func() {
			It("should return boolean indicating if host is resolvable", func() {
				it := New(`function FindProxyForURL(url, host) { return isResolvable(host); }`)
				Expect(it.ProxyFor("http://google.com")).Should(Equal("true"))
				Expect(it.ProxyFor("http://blah.blah.gobble")).Should(Equal("false"))
			})
		})

		Describe("shExpMatch", func() {
			It("should return boolean indicating if expression matches", func() {
				it := New(`function FindProxyForURL(url, host) { return shExpMatch(url, "*google.com*"); }`)
				Expect(it.ProxyFor("http://google.com/test")).Should(Equal("true"))
				Expect(it.ProxyFor("google.com/test")).Should(Equal("true"))
				Expect(it.ProxyFor("goggles.com")).Should(Equal("false"))

				it = New(`function FindProxyForURL(url, host) { return shExpMatch(url, "*google.com"); }`)
				Expect(it.ProxyFor("http://google.com/test")).Should(Equal("false"))
				Expect(it.ProxyFor("google.com")).Should(Equal("true"))
				Expect(it.ProxyFor("goggles.com")).Should(Equal("false"))

				it = New(`function FindProxyForURL(url, host) { return shExpMatch(url, "*goo?le.com*"); }`)
				Expect(it.ProxyFor("http://google.com/test")).Should(Equal("true"))
				Expect(it.ProxyFor("goodle.com")).Should(Equal("true"))
				Expect(it.ProxyFor("goggles.com")).Should(Equal("false"))
			})
		})

		Describe("isInNet", func() {
			It("should return boolean indicating if ip is in range", func() {
				it := New(`function FindProxyForURL(url, host) { return isInNet(url, "127.0.0.0", "255.255.255.0"); }`)
				Expect(it.ProxyFor("127.0.0.1")).Should(Equal("true"))
				Expect(it.ProxyFor("127.0.77.1")).Should(Equal("false"))
				Expect(it.ProxyFor("192.0.0.1")).Should(Equal("false"))

				it = New(`function FindProxyForURL(url, host) { return isInNet(url, "127.0.0.0", "255.255.0.0"); }`)
				Expect(it.ProxyFor("127.0.0.1")).Should(Equal("true"))
				Expect(it.ProxyFor("127.0.77.1")).Should(Equal("true"))
				Expect(it.ProxyFor("192.0.0.1")).Should(Equal("false"))
			})
		})

		Describe("isPlainHostName", func() {
			It("should return boolean indicating if hostname is plain", func() {
				it := New(`function FindProxyForURL(url, host) { return isPlainHostName(host); }`)
				Expect(it.ProxyFor("http://google.com")).Should(Equal("false"))
				Expect(it.ProxyFor("http://localhost")).Should(Equal("true"))
				Expect(it.ProxyFor("http://cheesesticks")).Should(Equal("true"))
			})
		})

		//PIt("should provide myIpAddress")
		//PIt("should provide localHostOrDomainIs")
		//PIt("should provide dnsDomainLevels")
		//PIt("should provide weekdayRange")
		//PIt("should provide dateRange")
		//PIt("should provide timeRange")
		//PIt("should provide alert")
	})
})

package proxyfactory_test

import (
	. "../../pacyak/proxyfactory"

	"net"
	"net/http"
	"net/url"

	"github.com/elazarl/goproxy"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Proxyfactory", func() {
	Describe("New", func() {
		It("should return instance of ProxyFactory", func() {
			Expect(New()).Should(BeAssignableToTypeOf(&ProxyFactory{}))
		})
	})

	Describe("Proxy", func() {
		It("should return a proxy for given handle", func() {
			it := New()
			Expect(it.Proxy("http://test.proxy")).Should(BeAssignableToTypeOf(&goproxy.ProxyHttpServer{}))
		})

		It("should return different proxies for each handle", func() {
			it := New()
			p1 := it.Proxy("http://test1.proxy")
			p2 := it.Proxy("http://test2.proxy")
			Expect(p1).ShouldNot(Equal(p2))
		})

		It("should return the same proxy for subsequent invocations of the same handle", func() {
			it := New()
			p1 := it.Proxy("http://test1.proxy")
			p2 := it.Proxy("http://test1.proxy")
			Expect(p1).Should(Equal(p2))
		})

		It("should configure the proxy", func() {
			it := New()
			url, _ := url.Parse("http://test.proxy")
			proxy := it.Proxy(url.String())
			Expect(proxy.Tr.Proxy(&http.Request{})).Should(Equal(url))
			// TODO who the fark to test ConnectDial?
		})
	})

	Describe("FromPacResponse", func() {
		It("should return a direct proxy if response is DIRECT", func() {
			factory := New()
			proxy := factory.FromPacResponse("DIRECT")
			var nilURL *url.URL
			var nilDial func(string, string) (net.Conn, error)
			Expect(proxy.Tr.Proxy(&http.Request{})).Should(Equal(nilURL))
			Expect(proxy.ConnectDial).Should(Equal(nilDial))
		})

		It("should return first proxy that is available", func() {
		})

		It("should return direct proxy if no proxy available", func() {
		})
	})
})

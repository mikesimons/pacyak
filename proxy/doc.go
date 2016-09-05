package proxy

/*
This proxy package is largely derived from github.com/elazarl/go-proxy so credit for much of this package belongs to elazarl.

Originally pacyak used go-proxy but this was like using a sledgehammer to crack a nut.
I also noticed that there were some issues around HTTPS requests getting dropped on the floor.

Using go-proxy as a reference we wrote this much simplified lib to do the proxying for pacyak.
This seems to have resolved the HTTPS issues, eliminates a dep and makes the proxying part easier to understand.

Please note that this implementation had very specific usage in mind and as such is probably more useful as a reference than anything else.
There is no MITM or filtering logic. If you need that I'd recommend go-proxy.

Usage is simple:

	proxy := proxy.New("direct")
	http.ListenAndServe("127.0.0.1:8080", proxy)

You can now set `http_proxy` and `https_proxy` to `http://127.0.0.1:8080` and all requests will be proxied.
This isn't particularly useful in itself because this proxy implementation had a very specific purpose; proxy to an upstream proxy...

	proxy := proxy.New("http://my-proxy.com")
	http.ListenAndServe("127.0.0.1:8080", proxy)

All requests going through the local proxy will be sent on to the upstream proxy.
*/

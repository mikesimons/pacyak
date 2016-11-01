# PacYak [![Build Status](https://travis-ci.org/mikesimons/pacyak.svg?branch=master)](https://travis-ci.org/mikesimons/pacyak)

 > For the unfortunate souls stuck behind corporate proxies

PacYak is a local proxy to assist developers behind corporate firewalls.

90% of connectivity problems I get asked to help with (and then even some problems that don't look like connectivity) stem from either proxies being set where they shouldn't or not being set where they should.
Typically corporate IT have codified a fairly decent set of rules for browsers to follow to decide which proxies to use and when in the form of a PAC file.

PacYak leverages this PAC file (or a user supplied one) to forward any traffic sent to it to the right place.

PacYak also has connectivity checks for both proxy and the auto config URL and falls back on direct connections when they are not available. This fixes the problems of taking a corporate laptop home and having to mess with proxy settings because you're no longer behind one (until you get back to the office at least).

PacYak can also be deployed inside VMs to leverage the PAC rules with CLI clients not directly supporting auto_proxy (but supporting http_proxy / https_proxy).

Using PacYak you can set your proxies and auto configuration urls to http://127.0.0.1:8080 (by default) and forget about them.

Please note that this has only been tested by me on one corporate network. It has worked well but YMMV.

## Limitations

- No daemon / service scripts so you have to figure out how to run it on startup yourself
- No support for authenticated proxies
- No PAC auto-discovery

## Installation

Pacyak needs to constantly be running. For now we're only providing pre-compiled binaries, not the necessary configuration to start Pacyak automatically so you'll have to take care of that yourself.

I also have not implemented WPAD or any other discovery as the network I'm on doesn't use it. If you can provide access to such an environment I will happily add support (or accept a PR for the same). Same thing for authenticated proxies.

That out of the way, here is how you start it:

```
pacyak http://my-corporate-proxy-pac-url:1234
```

Pacyak will fetch the PAC file at the URL and begin listening.
You should now configure your machine to use pacyak. You should probably start by making sure that pacyak is working as expected in a terminal with the following variables:

```
export HTTP_PROXY=127.0.0.1:8080
export HTTPS_PROXY=127.0.0.1:8080
export http_proxy=127.0.0.1:8080
export https_proxy=127.0.0.1:8080
```

Attempt to `curl` a page that requires a proxy and confirm that it returns the expected response.

If this works, you should set those environment variables on a wider basis. I don't have a reference for how to do this on all platforms but `/etc/environment` is a good bet for linux and maybe OSX and `Control Panel -> System -> Advanced Settings -> Environment Variables` for Windows.

If you're using a corporate laptop you may find that your browser proxy settings have also been set. You should change these for the values above.

## Troubleshooting
### Halp! It doesn't work!
Try turning up the log level with `--log-level debug` if you encounter problems. Errors should be reported at any reporting level but it might highlight an edge case / incompatibility I haven't considered.

### I need a proxy to get to the PAC file! How?
Use the `--pac-proxy` option to tell pacyak the proxy to use. This might seem crazy but the test network requires this when on VPN!

### IT are crazy / lazy and the PAC file is full of ascii cows. How can I use a local pac file?
Just create it locally and specify the path to it for the PAC location argument. You will also need to provide a host that is only accessible from within the proxy network via `--ping-host`. If this host is available globally pacyak will never switch to *not* using a proxy.

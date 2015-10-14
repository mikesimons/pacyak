# PacYak

PacYak is a local proxy designed to assist developers behind corporate firewalls.

90% of connectivity problems I get asked to resolve (and then even some problems that don't look like connectivity) stem from either proxies being set where they shouldn't or not being set where they should.
Typically corporate IT have codified a fairly decent set of rules for browsers to follow to decide which proxies to use and when in the form of a PAC file.

PacYak leverages this PAC file (or a user supplied one) to forward any traffic sent to it to the right proxy.

PacYak also has connectivity checks for both proxy and the auto config URL and falls back on direct connections when they are not available.
This fixes the problems of taking a corporate laptop home and having to mess with proxy settings because you're no longer behind one (until you get back to the office at least).

PacYak can also be deployed inside VMs to leverage the PAC rules with CLI clients not directly supporting auto_proxy (but supporting http_proxy / https_proxy).

Using PacYak you can set your proxies and auto configuration urls to http://127.0.0.1:8080 (by default) and forget about them.

Please note that this has only been tested by me on one corporate network. It has worked well although I suspect there may be some occasional hiccups with https responses getting lost. YMMV.

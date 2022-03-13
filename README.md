# LocalRelay

A cross platform CLI & lib which acts as a reverse proxy allowing the destination address to be customised and allows the use of a SOCKS5 proxy. Supporting both raw TCP connections and HTTP/HTTPS connections with options such as; IP locking, Certificate pinning. This app allows you to host services e.g. Nextcloud on Tor and access it on your mobile or laptop anywhere.

# Use Cases

If you self host a service for example; Bitwarden, Nextcloud, Syncthing, Graphana, Gitea... you may not want to expose your public IP address to the internet. Especially considering some self-hosted platforms such as Plex has been [exploited with code execution vulnerabilities](https://www.cvedetails.com/vulnerability-list.php?vendor_id=14994). You may consider to protect it behind Tor (however this isn't full proof).

Access your local services securely over [Tor](https://www.torproject.org/) without needing to port forward.

Many apps such as Nextcloud, Termis and Bitwarden do not allow you to specify a proxy when connecting to your self-hosted server. Localrelay allows you to host a local reverse proxy on your devices loopback. This relay then encrypts the outgoing traffic through your set SOCKS5 proxy (Tor: 127.0.0.1:9050).

# This Repository

This repository contains the library written in Go, for it's cross platform capabilities, and contains the CLI application which can be ran on all major operating systems including [Android via Termux](https://termux.com/).

For examples of API usage visit [examples/](https://github.com/go-compile/localrelay/tree/master/examples).

# Library Features

- Create relays with custom remote address
- Proxy remote address through SOCKS5 proxy
- Close relay concurrently
- Verbose logging with custom output (io.Writer)
- Metrics
  - Upload/Download
  - Total connections
  - Active connections
  - Dialler: successes/failures
  - Concurrent safe
  - Dialler 10 point average response time
    - When using Tor this is the circuit build time

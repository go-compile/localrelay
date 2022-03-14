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
- Multiple failover proxies for TCP relay
- HTTP relay
  - Http to https
  - Header modification
  - Useragent spoofing
  - Accept language spoofing
  - Proxy using socks5
- Metrics
  - Upload/Download
  - Total connections
  - Active connections
  - Dialler: successes/failures
  - Concurrent safe
  - Dialler 10 point average response time
    - When using Tor this is the circuit build time

# Privacy Proxies

Proxy your services whilst stripping personal information such as User-Agent, accept language or even cookies. Route the traffic through Tor to access the service anywhere in the word even behind a firewall.

![Relay spoofing useragent & using Tor](/examples/http-privacy/access-tor.png)

![Relay spoofing useragent & accept language](/examples/http-privacy/ifconfig.me.png)

# CLI Usage

You can download the CLI from the [releases tab](https://github.com/go-compile/localrelay/releases) or compile it your self by building [./cmd/localrelay](https://github.com/go-compile/localrelay/tree/cli.v1.0.0-alpha/cmd/localrelay). All releases hashed with SHA256 and signed.

Once you've downloaded the CLI you will need to give it execute permission if you're on a Unix based system. This is done with `chmod +x localrelay`. You don't need root permission to run the relay nor should you use it even if you want to run on a privileged port. Use `sudo setcap CAP_NET_BIND_SERVICE=+eip /path/to/localrelay` instead.

## Create Relay

To run a relay you must first create a relay config, this allows for permanent storage of relay configuration and easy management. You can create as many of these as you like.

### Syntax

```bash
# Create a simple TCP relay
localrelay new <relay_name> -host <bind_addr> -destination <remote_addr>

# Create HTTP relay
localrelay new <relay_name> -host <bind_addr> -destination <remote_addr> -http

# Create HTTPS relay
localrelay new <relay_name> -host <bind_addr> -destination <remote_addr> -https -certificate=cert.pem key=key.pem

# Use proxy
localrelay new <relay_name> -host <bind_addr> -destination <remote_addr> -proxy <proxy_url>

# Set custom output config file
localrelay new <relay_name> -host <bind_addr> -destination <remote_addr> -output ./config.toml
```

### Examples

```bash
# Create a simple TCP relay
localrelay new example.com -host 127.0.0.1:8080 -destination example.com:80

# Create HTTP relay
localrelay new example.com -host 127.0.0.1:8080 -destination http://example.com -http

# Create HTTPS relay
localrelay new example.com -host 127.0.0.1:8080 -destination https://example.com -https -certificate=cert.pem key=key.pem

# Use proxy
localrelay new onion -host 127.0.0.1:8080 -destination 2gzyxa5ihm7nsggfxnu52rck2vv4rvmdlkiu3zzui5du4xyclen53wid.onion:80 -proxy socks5://127.0.0.1:9050
```

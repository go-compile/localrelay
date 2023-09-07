# LocalRelay

[![GitHub release](https://img.shields.io/github/release/go-compile/localrelay.svg)](https://github.com/go-compile/localrelay/releases)
[![Go Report Card](https://goreportcard.com/badge/go-compile/localrelay)](https://goreportcard.com/report/go-compile/localrelay)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/go-compile/localrelay)
[![Docker Size](https://img.shields.io/docker/image-size/gocompile/localrelay?sort=date)](https://hub.docker.com/r/gocompile/localrelay/)
[![Docker Version](https://img.shields.io/docker/v/gocompile/localrelay?label=docker%20version&sort=semver)](https://hub.docker.com/r/gocompile/localrelay/)
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/go-compile/localrelay/.github/workflows/go.yml)

A cross platform CLI & lib which acts as a reverse proxy allowing the destination address to be customised and allows the use of a SOCKS5 proxy. Supporting both raw TCP connections and HTTP/HTTPS connections with options such as; IP locking, Certificate pinning. This app allows you to host services e.g. Nextcloud on Tor and access it on your mobile or laptop anywhere.

<div align=center>

**[\[ Wiki & Guide \]](https://github.com/go-compile/localrelay/wiki)**
[\[ Download Release \]](https://github.com/go-compile/localrelay/releases/latest)
[\[ Docker Image \]](https://hub.docker.com/r/gocompile/localrelay)

</div>

## Use Cases

If you self host a service for example; [Bitwarden](https://github.com/dani-garcia/vaultwarden), [Nextcloud](https://github.com/nextcloud), [Syncthing](https://github.com/syncthing/syncthing), [Grafana](https://github.com/grafana/grafana), [Gitea](https://github.com/go-gitea/gitea)... You may not want to expose your public IP address to the internet. Especially considering some self-hosted platforms such as [Plex](https://www.plex.tv/) has been exploited with code execution [vulnerabilities](https://www.cvedetails.com/vulnerability-list.php?vendor_id=14994). You may consider to protect it behind Tor (however this isn't full proof).

Access your local or remote services securely over [Tor](https://www.torproject.org/) without needing to port forward.

Many apps such as Nextcloud, Termis and Bitwarden do not allow you to specify a proxy when connecting to your self-hosted server. Localrelay allows you to host a local reverse proxy on your devices loopback. This relay then encrypts the outgoing traffic through your set SOCKS5 proxy (Tor: 127.0.0.1:9050).

When at **home connect locally**, when away **connect over Tor**. Securely connect remotely over Tor without port forwarding AND when at home connect directly with high speeds.

## This Repository

This repository contains the library written in Go, for it's cross platform capabilities, and contains the CLI application which can be ran on all major operating systems including [Android via Termux](https://termux.com/).

For examples of API usage visit [examples/](https://github.com/go-compile/localrelay/tree/master/examples).

## Library Features

Min Go version: `v1.17`
- Create relays with custom remote address
- Proxy remote address through SOCKS5 proxy
- Close relay concurrently
- Verbose logging with custom output (io.Writer)
- Multiple failover proxies for TCP relay
- Failovers for TCP relays
- Select which remote will connect via a proxy
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

## Privacy Proxies

Proxy your services whilst stripping personal information such as User-Agent, accept language or even cookies. Route the traffic through Tor to access the service anywhere in the word even behind a firewall.

<div align="center">

![Relay spoofing useragent & using Tor](/examples/http-privacy/access-tor.png)

![Relay spoofing useragent & accept language](/examples/http-privacy/ifconfig.me.png)

</div>

## CLI Usage

This is a basic overview, [view the wiki for more detailed information](https://github.com/go-compile/localrelay/wiki/CLI).

### Create Relay

To run a relay you must first create a relay config, this allows for permanent storage of relay configuration and easy management. You can create as many of these as you like.

#### Syntax

```sh
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

# Create a failover TCP relay
localrelay new <relay_name> -host <bind_addr> -destination <remote_addr_(1)>,<remote_addr_(2)> -failover
```

#### Examples

```sh
# Create a simple TCP relay
localrelay new example.com -host 127.0.0.1:8080 -destination example.com:80

# Create HTTP relay
localrelay new example.com -host 127.0.0.1:8080 -destination http://example.com -http

# Create HTTPS relay
localrelay new example.com -host 127.0.0.1:8080 -destination https://example.com -https -certificate=cert.pem key=key.pem

# Create a TCP relay and store it in the config dir to auto start on system boot (daemon required)
sudo localrelay new example.com -host 127.0.0.1:8080 -destination example.com:80 -store

# Use proxy
localrelay new onion -host 127.0.0.1:8080 -destination 2gzyxa5ihm7nsggfxnu52rck2vv4rvmdlkiu3zzui5du4xyclen53wid.onion:80 -proxy socks5://127.0.0.1:9050

# Create a failover TCP relay with one remote accessed via Tor
localrelay new onion -host 127.0.0.1:8080 -destination 192.168.1.240:80,2gzyxa5ihm7nsggfxnu52rck2vv4rvmdlkiu3zzui5du4xyclen53wid.onion:80 -failover -ignore_proxy=0 -proxy socks5://127.0.0.1:9050
```

<div align="center">

> localrelay status

![Localrelay status](.github/images/service.status.png)

</div>

<div align="center">

> localrelay monitor

![Localrelay status](.github/images/monitor.png)

</div>

<div align="center">
<br>

**[Installation And Usage Guide On The Wiki](https://github.com/go-compile/localrelay/wiki)**

</div>
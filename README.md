# Simple DNS reverse proxy

dnsrp allows you to easily configure which name server we should use based on domain.

## Install

`go get github.com/heppu/dnsrp`

## Usage

`dnsrp -c my-config.toml`

## Example use case

Let's say I use wlan for internet connection and cable to connect office network. I can easily create rules that for IPs in office network we use cable and for others IPs we use wlan. This would be fine if we only accessed resources with IP's and not with domains. This is where dnsrp steps in.

Let's say we have name server in our intranet with IP 192.168.1.11 and internal resources under int.mycomp.com. The dnsrp config for this setup would look like this:

```toml
defaultServer = "8.8.8.8"

[servers]
  "192.168.1.11" = [
    "int.mycomp.com",
  ]
```

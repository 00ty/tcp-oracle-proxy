# tcp-oracle-proxy

A minimal yet production-hardened TCP forwarding proxy designed for **Oracle databases** (port 1521).  
Keeps client sessions **alive across server restarts, network hiccups or fail-overs**—no application re-connection required.

---

## Features

| Feature | Status |
|---------|--------|
| Transparent TCP bridge | ✅ |
| Auto-reconnect with exponential back-off | ✅ |
| Buffer pool (64 KiB) for high throughput | ✅ |
| Graceful shutdown (waits for idle tunnels) | ✅ |
| Real-time tunnel counter | ✅ |
| **Zero external dependencies** | ✅ |

---

## Quick Start

```bash
# 1. Clone
git clone https://github.com/YOUR_GITHUB/tcp-oracle-proxy.git
cd tcp-oracle-proxy

# 2. Run
go run tcp_oracle_proxy.go

# 3. Connect your Oracle client to
localhost:9001

```
Traffic is forwarded to the default upstream 192.168.8.11:1521.
Edit the constants at the top of the file to change listen or target addresses.

## Build & Install
```bash
go build -ldflags="-s -w" -o tcp-oracle-proxy tcp_oracle_proxy.go
sudo cp tcp-oracle-proxy /usr/local/bin/
```

## Flags / Environment (planned)
Next releases will support:

- -l listen address
- -t target address
- -metrics Prometheus endpoint
- -retry-max max back-off duration
Until then, simply edit the const block and re-compile.

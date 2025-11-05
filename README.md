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

# Statusly

**Beautiful, lightweight self-hosted status page & uptime monitor.**

Single binary · SQLite · Modern UI · Docker ready

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Statusly monitors your websites, APIs, TCP services, and hosts and turns the results into a public status page, incident history, badges, and machine-readable metrics.

## Features

- 🎨 Public status page with dark / light mode
- ⚡ HTTP, TCP & ICMP monitoring
- 📊 24h uptime percentage + average/latest latency
- 📋 Incident history
- 🔔 Notifications: Telegram / Discord / Webhook
- 🏷️ SVG status badges
- 📦 Single binary with SQLite storage
- 🚀 Docker & docker-compose support
- 🔗 REST API + Prometheus metrics
- 📤 CSV export for recent check data
- 🔄 Auto-refreshing web UI

## Quick Start

### Docker (recommended)

```bash
git clone https://github.com/nasa2468/statusly.git
cd statusly
docker compose up -d --build
```

Open **http://localhost:8080**.

### Run with Go

```bash
git clone https://github.com/nasa2468/statusly.git
cd statusly
go mod tidy
go run ./cmd/statusly
```

## Configuration

Edit `config.yaml`:

```yaml
server:
  address: ":8080"

database: "data/statusly.db"

title: "My Services Status"
description: "Real-time status of my projects"

targets:
  - name: Website
    type: http
    address: https://example.com
    interval: 60
    timeout: 10

  - name: Database
    type: tcp
    address: 127.0.0.1:5432
    interval: 60
    timeout: 5

  - name: Gateway
    type: icmp
    address: 192.168.1.1
    interval: 60
    timeout: 5

notifications:
  - type: telegram
    token: "YOUR_BOT_TOKEN"
    chat_id: "YOUR_CHAT_ID"
    enabled: true

  - type: discord
    url: "https://discord.com/api/webhooks/..."
    enabled: true
```

> Keep notification credentials out of public repositories. Use environment-specific configuration or a secret manager when deploying Statusly publicly.

## API

| Endpoint | Description |
|----------|-------------|
| `GET /api/status` | Overall status + targets |
| `GET /api/incidents` | Recent failed checks |
| `GET /api/recent?limit=100` | Latest checks |
| `GET /api/history?target=Name&hours=24` | Latency history |
| `GET /api/export.csv?limit=1000` | Export recent checks as CSV |
| `GET /badge.svg` | Overall SVG badge |
| `GET /badge/{name}.svg` | Per-target SVG badge |
| `GET /metrics` | Prometheus metrics |
| `GET /healthz` | Health check |

All API endpoints above are read-only `GET` endpoints.

## Monitoring semantics

- HTTP responses with status codes **200–499** are considered reachable; a server returning `5xx` is considered down.
- TCP is considered up when a connection can be established.
- ICMP uses the system `ping` command and therefore requires `ping` to be available in the runtime environment.
- A target's uptime percentage is calculated from checks in the previous 24 hours.

## Roadmap

- [x] HTTP / TCP / ICMP monitoring
- [x] Notifications (Telegram, Discord, Webhook)
- [x] Status badges
- [x] Prometheus metrics
- [x] CSV export
- [ ] Maintenance windows
- [ ] Latency charts on dashboard
- [ ] Simple admin authentication
- [ ] Configurable data retention
- [ ] OpenPing integration

## Development

```bash
go test ./...
go build ./cmd/statusly
```

Pull requests and focused bug reports are welcome.

## License

MIT

---

Made with ❤️ by [nasa2468](https://github.com/nasa2468)

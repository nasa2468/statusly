# Statusly

**Beautiful, lightweight self-hosted status page & uptime monitor.**

Single binary · SQLite · Modern UI · Docker ready

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## Features

- 🎨 Beautiful public status page (dark / light mode)
- ⚡ HTTP, TCP & ICMP monitoring
- 📊 24h uptime percentage + latency
- 📋 Incident timeline
- 🔔 Notifications: Telegram / Discord / Webhook
- 🏷️ SVG status badges (perfect for README)
- 📦 Single binary, zero runtime dependencies
- 💾 SQLite storage
- 🚀 Docker & docker-compose support
- 🔗 REST API + Prometheus metrics
- 🔄 Auto-refresh every 30s

## Quick Start

### Docker (recommended)

```bash
git clone https://github.com/nasa2468/statusly.git
cd statusly
docker compose up -d --build
```

Open **http://localhost:8080**

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

## Status Badges

Overall status:

```markdown
![Status](https://your-domain.com/badge.svg)
```

Per target:

```markdown
![Website](https://your-domain.com/badge/Website.svg)
```

## API

| Endpoint | Description |
|----------|-------------|
| `GET /api/status` | Overall status + targets |
| `GET /api/incidents` | Recent incidents |
| `GET /api/recent?limit=100` | Latest checks |
| `GET /api/history?target=Name&hours=24` | Latency history |
| `GET /badge.svg` | Overall SVG badge |
| `GET /badge/{name}.svg` | Per-target SVG badge |
| `GET /metrics` | Prometheus metrics |
| `GET /healthz` | Health check |

## Roadmap

- [x] Notifications (Telegram, Discord, Webhook)
- [x] ICMP support
- [x] Status badges
- [ ] Maintenance windows
- [ ] Latency charts on dashboard
- [ ] Simple admin authentication
- [ ] OpenPing integration

## License

MIT

---

Made with ❤️ by [nasa2468](https://github.com/nasa2468)

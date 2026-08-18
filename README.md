# Statusly

**Beautiful, lightweight self-hosted status page & uptime monitor.**

Single binary · SQLite · Modern UI · Docker ready

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## Features

- 🎨 Beautiful public status page with dark/light mode
- ⚡ HTTP & TCP monitoring (ICMP coming soon)
- 📊 24h uptime percentage + latency
- 📋 Incident timeline
- 📦 Single binary, zero external dependencies at runtime
- 💾 SQLite storage (WAL mode)
- 🚀 Docker & docker-compose support
- 🔗 REST API + Prometheus metrics
- 🔄 Auto-refresh every 30 seconds

## Quick Start

### Option 1: Docker (recommended)

```bash
git clone https://github.com/nasa2468/statusly.git
cd statusly
docker compose up -d --build
```

Open **http://localhost:8080**

### Option 2: Run with Go

Requirements: Go 1.23+

```bash
git clone https://github.com/nasa2468/statusly.git
cd statusly
go mod tidy
go run ./cmd/statusly
```

Then open http://localhost:8080

## Configuration

Edit `config.yaml`:

```yaml
server:
  address: ":8080"

database: "data/statusly.db"

title: "My Services Status"
description: "Real-time status of my projects and infrastructure"

targets:
  - name: Website
    type: http
    address: https://example.com
    interval: 60        # seconds
    timeout: 10         # seconds

  - name: API
    type: http
    address: https://api.example.com/health
    interval: 30
    timeout: 5

  - name: Database
    type: tcp
    address: 127.0.0.1:5432
    interval: 60
    timeout: 5
```

Restart Statusly after changing the configuration.

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /api/status` | Overall status + all targets summary |
| `GET /api/incidents` | Recent down events |
| `GET /api/recent?limit=100` | Latest check results |
| `GET /api/history?target=Name&hours=24` | Latency history for charts |
| `GET /metrics` | Prometheus metrics |
| `GET /healthz` | Health check |

## Project Structure

```
statusly/
├── cmd/statusly/main.go      # Entry point
├── internal/
│   ├── api/                 # REST API + Prometheus
│   ├── checker/             # HTTP / TCP probes
│   ├── config/              # YAML config
│   └── storage/             # SQLite
├── web/index.html          # Status page frontend
├── config.yaml             # Example config
├── Dockerfile
└── docker-compose.yml
```

## Roadmap

- [ ] Notifications (Telegram, Discord, Email, Webhook)
- [ ] ICMP / Ping support
- [ ] Maintenance windows
- [ ] Embeddable status badge for README
- [ ] Simple admin authentication
- [ ] OpenPing integration
- [ ] Latency charts on the dashboard

## License

MIT

---

Made with ❤️ by [nasa2468](https://github.com/nasa2468)

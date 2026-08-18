# Statusly

**Beautiful, lightweight self-hosted status page & uptime monitor.**

Single binary · SQLite · Modern UI · Docker ready

## Features

- 🎨 Beautiful public status page (dark/light mode)
- ⚡ HTTP / TCP / ICMP monitoring
- 📊 Uptime percentage & latency charts
- 📋 Incident timeline
- 📦 Single binary (no runtime dependencies)
- 💾 SQLite storage
- 🚀 Docker & docker-compose support
- 🔗 REST API + Prometheus metrics
- 🔒 Optional admin authentication
- 🔄 Can work standalone or integrate with [OpenPing](https://github.com/nasa2468/openping)

## Quick Start

### With Docker (recommended)

```bash
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  --name statusly \
  ghcr.io/nasa2468/statusly:latest
```

Open http://localhost:8080

### With Go

```bash
go install github.com/nasa2468/statusly/cmd/statusly@latest
statusly
```

## Configuration

Create `config.yaml`:

```yaml
server:
  address: ":8080"
  # password: "your-admin-password"  # optional

database: "data/statusly.db"

title: "My Services Status"
description: "Current status of my projects and services"

targets:
  - name: Website
    type: http
    address: https://example.com
    interval: 60
    timeout: 10

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

## API

- `GET /api/status` – Overall status + targets
- `GET /api/incidents` – Recent incidents
- `GET /api/history?target=xxx&hours=24` – Latency history
- `GET /metrics` – Prometheus metrics
- `GET /healthz` – Health check

## Roadmap

- [ ] Notifications (Telegram, Discord, Email, Webhook)
- [ ] Maintenance windows
- [ ] Custom domains & branding
- [ ] Embeddable status badge
- [ ] Multi-user / team support
- [ ] OpenPing integration

## License

MIT

---

Made with ❤️ by [nasa2468](https://github.com/nasa2468)

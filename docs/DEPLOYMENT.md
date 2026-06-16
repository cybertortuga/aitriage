# AITriage Enterprise Deployment Guide

This document outlines the deployment strategy and requirements for AITriage Enterprise in a production environment.

## System Requirements

- **OS**: Linux (Debian 12 Bookworm or compatible recommended)
- **Docker**: Docker Engine 24.0.0+ and Docker Compose v2.20+
- **CPU**: 2 vCPUs minimum (4+ recommended for parallel scanning)
- **RAM**: 4GB minimum (8GB+ recommended)
- **Storage**: SSD for SQLite WAL performance (10GB+ free space)

## Quick Start (Production)

The quickest way to start AITriage Enterprise securely in production is using the `make` utility:

```bash
# 1. Clone the repository
git clone https://github.com/cybertortuga/aitriage.git
cd aitriage

# 2. Build and start the production stack
make enterprise-up
```

This will run `docker compose up -d` targeting the production-ready multi-stage Dockerfile. The Web UI and API will be available at `http://localhost:8080`.

## Architecture Overview

AITriage Enterprise is packaged as a single standalone Go binary containing the embedded Web UI, deployed alongside scanning tools (semgrep, bandit, trivy, gitleaks) inside a Debian slim container. 

The deployment relies on **Docker Volumes** for persistence:
- **`aitriage-data`**: Stores the `aitriage.db` SQLite database.

## Environment Variables

The following environment variables can be configured in `.env` or passed directly to Docker Compose:

| Variable | Default | Description |
|---|---|---|
| `JWT_SECRET` | `aitriage_default_secret...` | **CRITICAL**: The secret key for signing JWT auth tokens. Must be changed in production. |
| `DB_PATH` | `/app/data/aitriage.db` | The internal path to the SQLite database. |
| `LOG_LEVEL` | `info` | Logging verbosity (`debug`, `info`, `warn`, `error`). |
| `GEMINI_API_KEY` | (empty) | LLM Provider API Key for AI remediation analysis. |
| `OPENAI_API_KEY` | (empty) | LLM Provider API Key. |
| `ANTHROPIC_API_KEY`| (empty) | LLM Provider API Key. |
| `PROJECT_PATH` | `.` (Current Dir) | The path on the host machine to mount into `/project` for scanning. |

## Production Best Practices

1. **Change JWT_SECRET**: Ensure `JWT_SECRET` is set to a secure, random string (e.g., `openssl rand -base64 32`).
2. **Reverse Proxy**: Place AITriage behind a reverse proxy (Nginx, Traefik, or Azure Front Door) to handle SSL/TLS termination. 
3. **Volume Backups**: Regularly backup the Docker volume containing the SQLite database. SQLite operates in WAL mode; use tools like `sqlite3 .backup` or stop the container before raw copying.
4. **Health Checks**: The `/api/health` endpoint verifies both web service availability and database connectivity. It is integrated into the Docker Compose `healthcheck`.

## Managing the Service

**Stop the stack:**
```bash
make down
```

**View logs:**
```bash
docker compose logs -f web
```

**Interactive Scan in Docker:**
```bash
make docker-tui PROJECT=/path/to/your/code
```

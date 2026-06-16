# ─── Stage 1: Build Web UI ───────────────────────────────────────────────────
FROM node:22-bookworm AS web-builder
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm install
COPY web/ ./
RUN npm run build

# ─── Stage 2: Build Go binary ─────────────────────────────────────────────────
FROM golang:1.25-bookworm AS go-builder
WORKDIR /app

# C deps for tree-sitter CGO
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential git \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Synchronize web assets into the Go binary build context
COPY --from=web-builder /web/dist /app/internal/server/ui/dist
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /aitriage ./cmd/aitriage

# ─── Stage 3: Runtime with all security tools ─────────────────────────────────
FROM debian:bookworm-slim

LABEL org.opencontainers.image.title="AITriage"
LABEL org.opencontainers.image.description="AI-powered security scanner — all tools included"
LABEL org.opencontainers.image.source="https://github.com/cybertortuga/aitriage"

# System deps + runtime C libs (merged into single layer for cache)
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates curl git python3 python3-pip python3-venv \
    libgcc-s1 libc6 \
    && rm -rf /var/lib/apt/lists/*

# ── semgrep + bandit via pipx ─────────────────────────────────────────────────
RUN pip3 install --break-system-packages pipx && \
    pipx install semgrep && \
    pipx install bandit
ENV PATH="/root/.local/bin:${PATH}"

# ── gitleaks v8.30.1 ──────────────────────────────────────────────────────────
RUN ARCH="$(dpkg --print-architecture)" && \
    if [ "$ARCH" = "amd64" ]; then GL_ARCH="x64"; else GL_ARCH="$ARCH"; fi && \
    curl -sSfL "https://github.com/gitleaks/gitleaks/releases/download/v8.30.1/gitleaks_8.30.1_linux_${GL_ARCH}.tar.gz" \
    | tar -xz -C /usr/local/bin gitleaks && \
    chmod +x /usr/local/bin/gitleaks

# ── trivy v0.70.0 ────────────────────────────────────────────────────────────
RUN ARCH="$(dpkg --print-architecture)" && \
    if [ "$ARCH" = "amd64" ]; then TRIVY_ARCH="64bit"; elif [ "$ARCH" = "arm64" ]; then TRIVY_ARCH="ARM64"; fi && \
    curl -sSfL "https://github.com/aquasecurity/trivy/releases/download/v0.70.0/trivy_0.70.0_Linux-${TRIVY_ARCH}.tar.gz" \
    | tar -xz -C /usr/local/bin trivy && \
    chmod +x /usr/local/bin/trivy

WORKDIR /project

# GitHub Action entrypoint wrapper (referenced by action.yml via `entrypoint:`)
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# ── AITriage binary (LAST — changes every build, everything above is cached) ─
COPY --from=go-builder /aitriage /usr/local/bin/aitriage

# Health check for web mode
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s \
    CMD curl -f http://localhost:8080/api/health || exit 1

EXPOSE 8080

ENTRYPOINT ["aitriage"]
CMD ["web", "--port", "8080"]


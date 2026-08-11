# ─── Stage 1: Build Web UI ───────────────────────────────────────────────────
FROM node:22-bookworm AS web-builder
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# ─── Stage 2: Build Go binary ─────────────────────────────────────────────────
FROM golang:1.25.12-bookworm AS go-builder
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
ARG AITRIAGE_VERSION=dev
RUN CGO_ENABLED=1 go build -ldflags="-s -w -X main.Version=${AITRIAGE_VERSION}" -o /aitriage ./cmd/aitriage

# Build the latest upstream Gitleaks release with patched Go dependencies. The
# upstream v8.30.1 asset was built with Go 1.24.11 and x/crypto 0.35.0, both of
# which now have fixable HIGH CVEs. Module source is authenticated by Go sumdb.
FROM --platform=$BUILDPLATFORM golang:1.25.12-bookworm AS gitleaks-builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
RUN go mod download github.com/zricethezav/gitleaks/v8@v8.30.1 && \
    cp -a /go/pkg/mod/github.com/zricethezav/gitleaks/v8@v8.30.1/. /src/ && \
    chmod -R u+w /src && \
    go mod edit -require=golang.org/x/crypto@v0.52.0 && \
    go mod edit -require=golang.org/x/text@v0.39.0 && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS="$TARGETOS" GOARCH="$TARGETARCH" \
      go build -trimpath -ldflags="-s -w -X=github.com/zricethezav/gitleaks/v8/version.Version=v8.30.1" -o /gitleaks .

# Trivy v0.72.0 was released with Go 1.26.4 and oras-go 2.6.0. Rebuilding the
# exact tagged source on patched Go 1.26.5, oras-go 2.6.2, go-git 5.19.2,
# x/text 0.39.0, and gRPC-Go 1.82.1
# removes the fixable CVEs without changing Trivy's scanner version or behavior.
FROM --platform=$BUILDPLATFORM golang:1.26.5-bookworm AS trivy-builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
RUN go mod download github.com/aquasecurity/trivy@v0.72.0 && \
    cp -a /go/pkg/mod/github.com/aquasecurity/trivy@v0.72.0/. /src/ && \
    chmod -R u+w /src && \
    go mod edit -require=oras.land/oras-go/v2@v2.6.2 && \
    go mod edit -require=github.com/go-git/go-git/v5@v5.19.2 && \
    go mod edit -require=golang.org/x/text@v0.39.0 && \
    go mod edit -require=google.golang.org/grpc@v1.82.1 && \
    go mod tidy && \
    GOEXPERIMENT=jsonv2 CGO_ENABLED=0 GOOS="$TARGETOS" GOARCH="$TARGETARCH" \
      go build -trimpath -ldflags="-s -w -X github.com/aquasecurity/trivy/pkg/version/app.ver=0.72.0" -o /trivy ./cmd/trivy

# ─── Stage 3: Runtime with all security tools ─────────────────────────────────
FROM debian:bookworm-slim

LABEL org.opencontainers.image.title="AITriage"
LABEL org.opencontainers.image.description="AI-powered security scanner — all tools included"
LABEL org.opencontainers.image.source="https://github.com/dodobrands/aitriage"

# System deps + runtime C libs (merged into single layer for cache)
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates curl git python3 python3-pip python3-venv \
    libgcc-s1 libc6 \
    && rm -rf /var/lib/apt/lists/*

# ── semgrep + bandit via pipx ─────────────────────────────────────────────────
ENV PIPX_HOME=/opt/pipx
ENV PIPX_BIN_DIR=/usr/local/bin
# Semgrep 1.170.1 pins mcp 1.23.3, which has three fixable HIGH CVEs.
# AITriage never exposes Semgrep's MCP server; nevertheless, keep the
# transitive package patched and prove CLI compatibility below and in E2E.
# pipx can retain superseded dist-info directories after an in-place transitive
# upgrade. The imported modules are patched below; stale metadata is removed so
# scanners cannot mistake those superseded records for active packages.
RUN pip3 install --break-system-packages 'pipx==1.8.0' && \
    pipx install 'semgrep==1.170.1' && \
    pipx runpip semgrep install --no-cache-dir \
      'mcp==1.28.1' 'msgpack==1.2.1' 'setuptools==83.0.0' && \
    pipx install 'bandit==1.9.4' && \
    pipx runpip bandit install --no-cache-dir 'setuptools==83.0.0' && \
    pipx upgrade-shared && \
    /opt/pipx/shared/bin/python -m pip install --upgrade \
      'setuptools==83.0.0' 'wheel==0.47.0' 'jaraco.context==6.1.2' && \
    semgrep --version && semgrep scan --help >/dev/null && \
    /opt/pipx/venvs/semgrep/bin/python -c \
      "import importlib.metadata as m; assert m.version('mcp') == '1.28.1'; assert m.version('msgpack') == '1.2.1'; assert m.version('setuptools') == '83.0.0'" && \
    /opt/pipx/venvs/bandit/bin/python -c \
      "import importlib.metadata as m; assert m.version('setuptools') == '83.0.0'" && \
    find /opt/pipx /usr/local/lib -type d \
      \( -name 'msgpack-1.1.2.dist-info' -o -name 'setuptools-70.3.0.dist-info' \) \
      -prune -print -exec rm -rf '{}' + && \
    test -z "$(find /opt/pipx /usr/local/lib -type d \
      \( -name 'msgpack-1.1.2.dist-info' -o -name 'setuptools-70.3.0.dist-info' \) \
      -print -quit)" && \
    bandit --version && \
    rm -rf /root/.cache/pip

# ── Patched source-built external scanners ───────────────────────────────────
COPY --from=gitleaks-builder /gitleaks /usr/local/bin/gitleaks
COPY --from=trivy-builder /trivy /usr/local/bin/trivy

# Build-only Python packaging tools are not needed at runtime. Removing the
# distro setuptools metadata also removes two fixable CVEs from the final image.
RUN apt-get purge -y \
      python3-setuptools python3-pip python3-venv python3.11-venv \
      python3-pip-whl python3-setuptools-whl && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/* /root/.cache

# GitHub Action entrypoint wrapper (referenced by action.yml via `entrypoint:`)
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# ── AITriage binary (LAST — changes every build, everything above is cached) ─
ARG AITRIAGE_VERSION=dev
ENV AITRIAGE_VERSION=${AITRIAGE_VERSION}
COPY --from=go-builder /aitriage /usr/local/bin/aitriage

# Create non-root user
RUN groupadd -g 1000 aitriage && \
    useradd -u 1000 -g aitriage -s /bin/bash -m aitriage && \
    mkdir -p /project && chown -R aitriage:aitriage /project

# Note: For GitHub Actions compatibility (writing to host-mounted GITHUB_WORKSPACE),
# we run as root by default. You can run as non-root locally using `docker run --user 1000`.
# USER aitriage
WORKDIR /project

EXPOSE 8080

ENTRYPOINT ["aitriage"]
CMD ["web", "--runtime", "native", "--port", "8080", "--host-prefix", "/workspace"]

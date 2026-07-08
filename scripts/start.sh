#!/bin/bash
# AITriage — One-command launcher for macOS
# Usage: ./start.sh [port]
set -e

PORT=${1:-8080}
IMAGE="ghcr.io/cybertortuga/aitriage:latest"
CONTAINER="aitriage-web"

echo ""
echo "  ╔══════════════════════════════════════╗"
echo "  ║        AITriage Web UI               ║"
echo "  ╚══════════════════════════════════════╝"
echo ""

# Stop old container if running
if docker ps -q -f name="$CONTAINER" | grep -q .; then
    echo "  → Stopping existing container..."
    docker stop "$CONTAINER" > /dev/null
fi
if docker ps -aq -f name="$CONTAINER" | grep -q .; then
    docker rm "$CONTAINER" > /dev/null
fi

# Pull latest image
echo "  → Pulling latest image..."
docker pull "$IMAGE" 2>/dev/null || echo "  ⚠ Could not pull, using local build..."
# Run
echo "  → Starting AITriage..."
docker run -d \
    --name "$CONTAINER" \
    -p "${PORT}:8080" \
    -v "${HOME}:${HOME}:ro" \
    -e OPENAI_API_KEY="${OPENAI_API_KEY:-}" \
    -e ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}" \
    --restart unless-stopped \
    "$IMAGE" web --port 8080 --host-prefix "" > /dev/null

echo ""
echo "  ✅ AITriage running at → http://localhost:${PORT}"
echo ""
echo "  Enter any project path on your Mac, e.g.:"
echo "  ${HOME}/Documents/GitHub/myproject"
echo ""
echo "  To stop:  docker stop $CONTAINER"
echo "  To logs:  docker logs -f $CONTAINER"
echo ""

# Open browser on Mac
if command -v open &> /dev/null; then
    sleep 1 && open "http://localhost:${PORT}" &
fi

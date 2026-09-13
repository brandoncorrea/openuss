#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROXY_CONF="$SCRIPT_DIR/localutm-proxy.conf"
PROXY_NAME=localutm-proxy

# Stop the proxy
docker rm -f "$PROXY_NAME" >/dev/null 2>&1 || true

# Restart Mocks
(
  cd "$INTERUSS_MONITORING_DIR"
  make restart-all
  ./monitoring/uss_qualifier/run_locally.sh configurations.dev.noop
)

# Start Proxy
docker run -d --name "$PROXY_NAME" \
  --network interop_ecosystem_network \
  -p 80:80 \
  -v "$PROXY_CONF:/etc/nginx/conf.d/default.conf:ro" \
  "nginx:1.27-alpine" >/dev/null
echo "$PROXY_NAME listening on :80 — add the /etc/hosts line from $PROXY_CONF if you haven't"

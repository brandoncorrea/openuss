#!/bin/bash

set -euo pipefail

CONFIG_DIR_NAME="openuss"
CONFIG_NAME="openuss_local"
DEFAULT_BASE_URL="http://openuss.uss5.localutm:8080"
DEFAULT_STOP_FAST="true"
USS_BASE_URL="${USS_BASE_URL:-$DEFAULT_BASE_URL}"
QUALIFIER_STOP_FAST="${QUALIFIER_STOP_FAST:-$DEFAULT_STOP_FAST}"

cd "$(dirname "${BASH_SOURCE[0]}")/.."

SRC_DIR="./test/config/utm_implementation_us"
DEST_DIR="$INTERUSS_MONITORING_DIR/monitoring/uss_qualifier/configurations/dev/utm_implementation_us/environments/$CONFIG_DIR_NAME"

mkdir -p "$DEST_DIR"
for src in "$SRC_DIR"/*; do
  sed -e "s|$DEFAULT_BASE_URL|$USS_BASE_URL|g" \
      -e "s|^\( *\)stop_fast: $DEFAULT_STOP_FAST,\$|\1stop_fast: $QUALIFIER_STOP_FAST,|" \
      "$src" > "$DEST_DIR/$(basename "$src")"
done

cd "$INTERUSS_MONITORING_DIR"
./monitoring/uss_qualifier/run_locally.sh "configurations.dev.utm_implementation_us.environments.$CONFIG_DIR_NAME.$CONFIG_NAME"

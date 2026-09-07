#!/bin/bash

set -euo pipefail

CONFIG_NAME="openuss_f3548"
DEFAULT_BASE_URL="http://openuss.uss5.localutm:8080"
DEFAULT_STOP_FAST="true"
USS_BASE_URL="${USS_BASE_URL:-$DEFAULT_BASE_URL}"
QUALIFIER_STOP_FAST="${QUALIFIER_STOP_FAST:-$DEFAULT_STOP_FAST}"

cd "$(dirname "${BASH_SOURCE[0]}")/.."

SRC="./test/config/$CONFIG_NAME.yaml"
DEST="$INTERUSS_MONITORING_DIR/monitoring/uss_qualifier/configurations/dev/$CONFIG_NAME.yaml"

sed -e "s|$DEFAULT_BASE_URL|$USS_BASE_URL|g" \
    -e "s|^\( *\)stop_fast: $DEFAULT_STOP_FAST\$|\1stop_fast: $QUALIFIER_STOP_FAST|" \
    "$SRC" > "$DEST"

cd $INTERUSS_MONITORING_DIR
./monitoring/uss_qualifier/run_locally.sh "configurations.dev.$CONFIG_NAME"

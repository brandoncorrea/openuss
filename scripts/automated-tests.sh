#!/bin/bash

set -euo pipefail

CONFIG_NAME="openuss_f3548"
DEFAULT_BASE_URL="http://openuss.uss5.localutm:8080"
USS_BASE_URL="${USS_BASE_URL:-$DEFAULT_BASE_URL}"

cd "$(dirname "${BASH_SOURCE[0]}")/.."

SRC="./test/config/$CONFIG_NAME.yaml"
DEST="$INTERUSS_MONITORING_DIR/monitoring/uss_qualifier/configurations/dev/$CONFIG_NAME.yaml"

sed "s|$DEFAULT_BASE_URL|$USS_BASE_URL|g" $SRC > $DEST

cd $INTERUSS_MONITORING_DIR
./monitoring/uss_qualifier/run_locally.sh "configurations.dev.$CONFIG_NAME"

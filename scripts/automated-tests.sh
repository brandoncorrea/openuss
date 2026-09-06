#!/bin/bash

set -euo pipefail

CONFIG_NAME="openuss_f3548"

cd "$(dirname "${BASH_SOURCE[0]}")/.."

cp "./test/config/$CONFIG_NAME.yaml" "$INTERUSS_MONITORING_DIR/monitoring/uss_qualifier/configurations/dev/"

cd $INTERUSS_MONITORING_DIR
./monitoring/uss_qualifier/run_locally.sh "configurations.dev.$CONFIG_NAME"

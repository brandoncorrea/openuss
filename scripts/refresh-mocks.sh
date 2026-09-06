#!/bin/bash

set -euo pipefail

cd $INTERUSS_MONITORING_DIR
make restart-all
./monitoring/uss_qualifier/run_locally.sh configurations.dev.noop

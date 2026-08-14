#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "$0")/.."
REPO_ROOT="$(pwd)"

DSS_REF="0642018eae0daf0562f5f891dfa3bd71605acf7c"
GENERATOR_IMAGE="openuss/openapi-to-go-server:${DSS_REF}"

API_FOLDER="${REPO_ROOT}/internal/api"
API_IMPORT="bwawan.com/openuss/internal/api"

PROTOCOL="${REPO_ROOT}/interfaces/astm-utm/Protocol"

if [ ! -f "${PROTOCOL}/utm.yaml" ]; then
  echo "interfaces/astm-utm/Protocol is missing. Run:" >&2
  echo "  git submodule update --init --recursive" >&2
  exit 1
fi

echo "==> Building ${GENERATOR_IMAGE}"
docker image build -q -t "${GENERATOR_IMAGE}" \
  "https://github.com/interuss/dss.git#${DSS_REF}:interfaces/openapi-to-go-server" >/dev/null

echo "==> Generating ${API_IMPORT}"
rm -rf "${API_FOLDER}"
mkdir -p "${API_FOLDER}"
docker container run --rm -u "$(id -u):$(id -g)" \
  -v "${PROTOCOL}:/spec/protocol:ro" \
  -v "${API_FOLDER}:/resources/internal/api" \
  "${GENERATOR_IMAGE}" \
  --api "/spec/protocol/utm.yaml#p2p_utm,dss@scdussv1" \
  --api_folder /resources/internal/api \
  --api_import "${API_IMPORT}"

echo "==> Removing server.gen.go"
find "${API_FOLDER}" -name 'server.gen.go' -delete

echo "==> Done"
find "${API_FOLDER}" -name '*.gen.go' | sort | sed "s|${REPO_ROOT}/|  |"

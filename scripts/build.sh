#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
. ./scripts/go-env.sh
(cd web && npm ci && npm run build)
mkdir -p dist/web
(cd server && GOWORK=off CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o ../dist/probelune-dashboard ./cmd/server)
cp -R web/dist/. dist/web/
cp demo/demo_seed.json dist/demo_seed.json

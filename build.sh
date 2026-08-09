#!/usr/bin/env sh
set -eu

APP_NAME="icloud-prime"
OUT_DIR="build"

mkdir -p "$OUT_DIR"

if [ -d web ]; then
  (
    cd web
    npm ci
    npm run build
  )
fi

CGO_ENABLED=0 go build -ldflags="-s -w" -o "$OUT_DIR/$APP_NAME" .

echo "Built $OUT_DIR/$APP_NAME"

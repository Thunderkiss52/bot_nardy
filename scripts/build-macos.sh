#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "error: this script must be run on macOS (Darwin)"
  exit 1
fi

if ! command -v go >/dev/null 2>&1; then
  echo "error: Go is not installed"
  exit 1
fi

if ! xcode-select -p >/dev/null 2>&1; then
  echo "error: Xcode Command Line Tools not found"
  echo "run: xcode-select --install"
  exit 1
fi

GOCACHE="${GOCACHE:-/tmp/go-cache}"
export GOCACHE

GOFLAGS="${GOFLAGS:-}"
export GOFLAGS

echo "[1/3] Running tests"
go test ./...

echo "[2/3] Building macOS app bundle"
go run github.com/wailsapp/wails/v2/cmd/wails@v2.11.0 build -tags wails -platform darwin/universal -clean

APP_PATH="build/bin/desktop-nardy-engine.app"
ZIP_PATH="build/bin/desktop-nardy-engine-macos.zip"

if [[ ! -d "$APP_PATH" ]]; then
  echo "error: build completed but app bundle not found at $APP_PATH"
  exit 1
fi

echo "[3/3] Packing zip artifact"
rm -f "$ZIP_PATH"
ditto -c -k --sequesterRsrc --keepParent "$APP_PATH" "$ZIP_PATH"

echo "done"
echo "app: $ROOT_DIR/$APP_PATH"
echo "zip: $ROOT_DIR/$ZIP_PATH"

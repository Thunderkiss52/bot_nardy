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

echo "[1/5] Running tests"
go test ./...

echo "[2/5] Building macOS app bundle"
go run github.com/wailsapp/wails/v2/cmd/wails@v2.11.0 build -tags wails -platform darwin/universal -clean

APP_PATH="build/bin/desktop-nardy-engine.app"
ZIP_PATH="build/bin/desktop-nardy-engine-macos.zip"
SIGNED_ZIP_PATH="build/bin/desktop-nardy-engine-notary.zip"

if [[ ! -d "$APP_PATH" ]]; then
  echo "error: build completed but app bundle not found at $APP_PATH"
  exit 1
fi

if [[ -n "${MACOS_SIGN_IDENTITY:-}" ]]; then
  echo "[3/5] Signing with Developer ID identity"
  codesign --force --deep --options runtime --timestamp --sign "$MACOS_SIGN_IDENTITY" "$APP_PATH"
else
  echo "[3/5] Ad-hoc signing (for local testing)"
  codesign --force --deep --sign - "$APP_PATH"
fi

codesign --verify --deep --strict --verbose=2 "$APP_PATH"
echo "[info] Gatekeeper assessment:"
spctl -a -vv "$APP_PATH" || true

if [[ -n "${MACOS_NOTARY_PROFILE:-}" ]]; then
  echo "[4/5] Notarizing app bundle"
  rm -f "$SIGNED_ZIP_PATH"
  ditto -c -k --sequesterRsrc --keepParent "$APP_PATH" "$SIGNED_ZIP_PATH"
  xcrun notarytool submit "$SIGNED_ZIP_PATH" --keychain-profile "$MACOS_NOTARY_PROFILE" --wait
  xcrun stapler staple "$APP_PATH"
  xcrun stapler validate "$APP_PATH"
else
  echo "[4/5] Notarization skipped (MACOS_NOTARY_PROFILE not set)"
fi

echo "[5/5] Packing zip artifact"
rm -f "$ZIP_PATH"
ditto -c -k --sequesterRsrc --keepParent "$APP_PATH" "$ZIP_PATH"

echo "done"
echo "app: $ROOT_DIR/$APP_PATH"
echo "zip: $ROOT_DIR/$ZIP_PATH"

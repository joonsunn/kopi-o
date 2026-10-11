#!/bin/sh
# package kopi-o into bin/Kopi-O.app (+ bin/Kopi-O.zip), ad-hoc signed.
# No sudo, no Xcode project. Requires: Go toolchain + CLT clang (cgo).
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
BIN="$ROOT/bin/kopi-o"
APP="$ROOT/bin/Kopi-O.app"
ZIP="$ROOT/bin/Kopi-O.zip"

mkdir -p "$ROOT/bin"
echo "==> build $BIN"
CGO_ENABLED=1 go build -o "$BIN" ./cmd/kopi-o

echo "==> assemble $APP"
rm -rf "$APP"
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources"
cp "$BIN" "$APP/Contents/MacOS/kopi-o"
cp "$ROOT/packaging/Info.plist" "$APP/Contents/Info.plist"

echo "==> ad-hoc codesign"
codesign --force --deep -s - "$APP"

echo "==> zip $ZIP"
rm -f "$ZIP"
(cd "$ROOT/bin" && zip -qr "Kopi-O.zip" "Kopi-O.app")

echo "packaged: $APP"
codesign -dv "$APP" 2>&1 | head -5

#!/bin/sh
# install Kopi-O.app to ~/Applications (user scope, no sudo).
#
#   ./scripts/install.sh                      # app only
#   INSTALL_LAUNCH_AGENT=1 ./scripts/install.sh  # + login item plist
#
# Dry run: DESTDIR=/tmp/dry-run ./scripts/install.sh
#   installs to $DESTDIR$HOME/Applications instead of $HOME/Applications.
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
APP_NAME="Kopi-O.app"
ZIP="$ROOT/bin/$APP_NAME"

DEST_HOME="${DESTDIR:-}$HOME"
APP_DIR="$DEST_HOME/Applications"
APP_DST="$APP_DIR/$APP_NAME"

if [ ! -d "$ROOT/bin/$APP_NAME" ]; then
	echo "missing $ROOT/bin/$APP_NAME — run ./scripts/package.sh (or make package) first" >&2
	exit 1
fi

echo "==> install $APP_DST"
mkdir -p "$APP_DIR"
rm -rf "$APP_DST"
cp -R "$ROOT/bin/$APP_NAME" "$APP_DST"

if [ "${INSTALL_LAUNCH_AGENT:-0}" = "1" ]; then
	PLIST_DIR="$DEST_HOME/Library/LaunchAgents"
	mkdir -p "$PLIST_DIR"
	sed "s|@EXEC@|$APP_DST/Contents/MacOS/kopi-o|g" \
		"$ROOT/packaging/LaunchAgent.plist" > "$PLIST_DIR/com.kopio.kopi-o.plist"
	echo "login item: $PLIST_DIR/com.kopio.kopi-o.plist"
	if [ -z "${DESTDIR:-}" ]; then
		echo "hint: log out/in (or run: launchctl bootstrap gui/$(id -u) $PLIST_DIR/com.kopio.kopi-o.plist)"
	fi
fi

echo "installed: $APP_DST"

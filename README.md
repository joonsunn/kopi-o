# kopi-o

macOS menu-bar app (Go) that holds idle sleep for a preset duration, Caffeine-style. User-local, no sudo. Build dep: Xcode CLT (+ Go toolchain).

- Tray: `getlantern/systray` (cgo). Sleep: `/usr/bin/caffeinate` child proc.
- Config: `$XDG_CONFIG_HOME/kopi-o/config.json` fallback `~/.config/kopi-o/config.json` (`KOPI_O_CONFIG` override).
- `make setup build test run package install` (see Makefile). Installs to `~/Applications`, no sudo.

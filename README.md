# kopi-o

macOS menu-bar app (Go) that holds idle sleep for a preset duration, Caffeine-style. User-local, no sudo. Build dep: Xcode CLT (+ Go toolchain).

- Tray: `fyne.io/systray` (cgo, pinned v1.12.2). Sleep: `/usr/bin/caffeinate` child proc.
- Config: `$XDG_CONFIG_HOME/kopi-o/config.json` fallback `~/.config/kopi-o/config.json` (`KOPI_O_CONFIG` override). 0700 dir, 0600 file, atomic writes. Missing/corrupt → 1h defaults.
- `make setup build test run package install` (see Makefile). Installs to `~/Applications`, no sudo.

## Menu

Left-click the tray icon toggles the hold (idle → start last preset, default 1h; active → stop). Right-click opens the menu.

Presets (radio): Off / 15m / 1h / 2h / 4h / ∞, plus Stop, Launch-at-login checkbox, Quit. Icon legend: filled cup + steam = holding, cup outline = idle. Title shows countdown only while holding (`42m`, `1h05m`, `38s`, `∞`; empty when idle). The selected preset persists across restarts and is restored on launch (Off stays idle).

## Install

```sh
make package   # bin/Kopi-O.app (LSUIElement, no Dock icon) + bin/Kopi-O.zip, ad-hoc signed
make install   # copy to ~/Applications
INSTALL_LAUNCH_AGENT=1 make install  # + login item (~/Library/LaunchAgents)
```

First launch: right-click `~/Applications/Kopi-O.app` → Open (ad-hoc signature needs one manual approval; no notarization in MVP). Or tick “Launch at login” in the menu to write the login item without reinstalling.

Dry run without touching home:

```sh
DESTDIR=/tmp/dry-run ./scripts/install.sh
```

## Uninstall

Quit kopi-o, then:

```sh
rm -rf ~/Applications/Kopi-O.app ~/Library/LaunchAgents/com.kopio.kopi-o.plist
```

(Config in `~/.config/kopi-o/` can stay or go.)

## Limitations

- Prevents idle sleep only; lid-close/clamshell and MDM screen-lock/policy still win.
- An indefinite hold left on drains battery — the filled-cup icon and `∞` title are prominent and Stop is one click.
- `pmset -g assertions | grep -i caffeinate` is the ground truth for “holding”.

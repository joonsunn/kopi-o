package tray

import (
	_ "embed"
	"fmt"
	"log"
	"sync"
	"time"

	"fyne.io/systray"

	"kopio/internal/config"
	"kopio/internal/sleep"
)

// Ticker interval for the countdown title refresh (within 5–15s spec).
const tickInterval = 10 * time.Second

// Preset is a selectable hold duration.
type Preset struct {
	Label      string
	Secs       int
	Indefinite bool
}

// Presets in menu order. Off (Secs 0) is the idle state.
var Presets = []Preset{
	{Label: "Off", Secs: 0},
	{Label: "15m", Secs: 900},
	{Label: "1h", Secs: 3600},
	{Label: "2h", Secs: 7200},
	{Label: "4h", Secs: 14400},
	{Label: "∞", Indefinite: true},
}

//go:embed icons/on.png
var iconOn []byte

//go:embed icons/off.png
var iconOff []byte

// Deps wires the tray to the supervisor. Sup must be non-nil;
// its onExpire callback should call Refresh (see cmd/kopi-o).
type Deps struct {
	Sup *sleep.Supervisor
}

var (
	mu        sync.Mutex
	deps      Deps
	presetIx  = map[*systray.MenuItem]Preset{}
	stopItem  *systray.MenuItem
	loginItem *systray.MenuItem
	ready     = make(chan struct{})
	// selected is the radio label reflecting user choice or restore.
	// Expiry/stop transitions reset it to "Off".
	selected  = "Off"
	wasActive = false
)

// Run blocks on the systray main loop. Call from main().
func Run(d Deps) {
	deps = d
	systray.Run(onReady, onExit)
}

func onReady() {
	close(ready)
	// Left-click toggles the hold; right-click still opens the menu
	// (tappedRight stays nil). Registered before the menu build.
	systray.SetOnTapped(toggle)
	cfg := config.Load()

	// Best-effort login-item sync with persisted config.
	if cfg.LaunchAtLogin {
		if err := SetLaunchAtLogin(true, ""); err != nil {
			log.Printf("tray: ensure launch agent: %v", err)
		}
	}

	// Restore persisted hold (Off = idle) and radio.
	if cfg.Indefinite {
		if err := deps.Sup.Start(0, true); err != nil {
			log.Printf("tray: restore indefinite hold: %v", err)
		} else {
			selected = "∞"
		}
	} else if cfg.PresetSecs > 0 {
		if err := deps.Sup.Start(cfg.PresetSecs, false); err != nil {
			log.Printf("tray: restore %ds hold: %v", cfg.PresetSecs, err)
		} else {
			selected = labelFor(cfg.PresetSecs, false)
		}
	}

	for _, p := range Presets {
		item := systray.AddMenuItem(p.Label, "Hold sleep for "+p.Label)
		presetIx[item] = p
		go watchPreset(item, p)
	}
	stopItem = systray.AddMenuItem("Stop", "Release the sleep hold")
	go watchStop()
	systray.AddSeparator()
	checked := cfg.LaunchAtLogin || IsLaunchAtLogin()
	loginItem = systray.AddMenuItemCheckbox("Launch at login", "Start kopi-o on login", checked)
	go watchLogin()
	quitItem := systray.AddMenuItem("Quit", "Quit kopi-o")
	go func() {
		<-quitItem.ClickedCh
		deps.Sup.Stop()
		systray.Quit()
	}()

	refreshLocked()
	go ticker()
}

func onExit() {
	deps.Sup.Stop()
}

func watchPreset(item *systray.MenuItem, p Preset) {
	for range item.ClickedCh {
		mu.Lock()
		if p.Secs == 0 && !p.Indefinite {
			deps.Sup.Stop()
			selected = "Off"
			persist(config.Config{LaunchAtLogin: loginChecked()})
		} else if err := deps.Sup.Start(p.Secs, p.Indefinite); err != nil {
			log.Printf("tray: start %s: %v", p.Label, err)
		} else {
			selected = p.Label
			persist(config.Config{
				PresetSecs:    p.Secs,
				Indefinite:    p.Indefinite,
				LaunchAtLogin: loginChecked(),
			})
		}
		mu.Unlock()
		Refresh()
	}
}

// toggle flips the hold on left-click: active → stop and reset to Off;
// idle → start the last persisted preset (default 1h when Off/empty).
// Same locking/persist pattern as watchPreset.
func toggle() {
	mu.Lock()
	if deps.Sup.Active() {
		deps.Sup.Stop()
		selected = "Off"
		persist(config.Config{LaunchAtLogin: loginChecked()})
	} else {
		p := lastPreset()
		if err := deps.Sup.Start(p.Secs, p.Indefinite); err != nil {
			log.Printf("tray: toggle start %s: %v", p.Label, err)
		} else {
			selected = p.Label
			persist(config.Config{
				PresetSecs:    p.Secs,
				Indefinite:    p.Indefinite,
				LaunchAtLogin: loginChecked(),
			})
		}
	}
	mu.Unlock()
	Refresh()
}

// lastPreset returns the hold a toggle-on starts: the persisted preset,
// defaulting to 1h when Off/empty.
func lastPreset() Preset {
	cfg := config.Load()
	if cfg.Indefinite {
		return Preset{Label: "∞", Indefinite: true}
	}
	if cfg.PresetSecs > 0 {
		if l := labelFor(cfg.PresetSecs, false); l != "" {
			return Preset{Label: l, Secs: cfg.PresetSecs}
		}
		return Preset{Label: fmt.Sprintf("%ds", cfg.PresetSecs), Secs: cfg.PresetSecs}
	}
	return Preset{Label: "1h", Secs: 3600}
}

func watchStop() {
	for range stopItem.ClickedCh {
		mu.Lock()
		deps.Sup.Stop()
		selected = "Off"
		persist(config.Config{LaunchAtLogin: loginChecked()})
		mu.Unlock()
		Refresh()
	}
}

func watchLogin() {
	for range loginItem.ClickedCh {
		on := !loginItem.Checked()
		if err := SetLaunchAtLogin(on, ""); err != nil {
			log.Printf("tray: launch at login: %v", err)
			continue
		}
		if on {
			loginItem.Check()
		} else {
			loginItem.Uncheck()
		}
		cfg := config.Load()
		cfg.LaunchAtLogin = on
		persist(cfg)
	}
}

func ticker() {
	t := time.NewTicker(tickInterval)
	defer t.Stop()
	for range t.C {
		Refresh()
	}
}

// Refresh recomputes the title, tooltip, and radio checks from supervisor
// state. Safe to call from any goroutine (expiry callback, ticker, clicks).
// Before systray ready it is a no-op.
func Refresh() {
	mu.Lock()
	defer mu.Unlock()
	select {
	case <-ready:
	default:
		return
	}
	refreshLocked()
}

// refreshLocked applies supervisor state to systray. Caller holds mu,
// systray must be ready.
func refreshLocked() {
	active := deps.Sup.Active()
	indef := deps.Sup.Indefinite()
	rem := deps.Sup.Remaining()

	// Expiry (or external stop) transitions the radio back to Off and
	// persist it so a restart doesn't re-hold a dead timer.
	if wasActive && !active {
		selected = "Off"
		persist(config.Config{LaunchAtLogin: loginChecked()})
	}
	wasActive = active

	systray.SetTitle(Title(active, indef, rem))
	if active {
		systray.SetTemplateIcon(iconOn, iconOn)
	} else {
		systray.SetTemplateIcon(iconOff, iconOff)
	}
	systray.SetTooltip(Tooltip(active, indef, rem, lastPreset().Label))

	current := selected
	if !active {
		current = "Off"
		selected = "Off"
	}
	for item, p := range presetIx {
		if p.Label == current {
			item.Check()
		} else {
			item.Uncheck()
		}
	}
}

// labelFor maps persisted secs to a radio label, "" when custom.
func labelFor(secs int, indefinite bool) string {
	if indefinite {
		return "∞"
	}
	for _, p := range Presets {
		if !p.Indefinite && p.Secs == secs {
			return p.Label
		}
	}
	return ""
}

// persist saves cfg, logging (not failing) on error: the tray stays usable
// with in-memory state if the disk write fails.
func persist(cfg config.Config) {
	if err := config.Save(cfg); err != nil {
		log.Printf("tray: save config: %v", err)
	}
}

func loginChecked() bool {
	return loginItem != nil && loginItem.Checked()
}

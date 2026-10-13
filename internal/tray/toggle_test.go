package tray

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"kopio/internal/config"
	"kopio/internal/sleep"
)

// TestToggle exercises the left-click code path headlessly: systray is never
// ready in tests, so Refresh is a no-op and no GUI is needed, while the real
// Supervisor spawns caffeinate against a temp config.
func TestToggle(t *testing.T) {
	t.Setenv("KOPI_O_CONFIG", filepath.Join(t.TempDir(), "config.json"))

	oldDeps, oldSelected := deps, selected
	defer func() { deps, selected = oldDeps, oldSelected }()

	deps = Deps{Sup: sleep.New(nil)}
	selected = "Off"
	me := os.Getpid()

	// Idle with empty config → toggle starts the default 1h hold.
	toggle()
	if !deps.Sup.Active() {
		t.Fatal("Sup.Active = false after toggle-on")
	}
	if selected != "1h" {
		t.Fatalf("selected = %q after toggle-on, want 1h", selected)
	}
	if got := config.Load(); got.PresetSecs != 3600 {
		t.Fatalf("persisted PresetSecs = %d, want 3600", got.PresetSecs)
	}
	if !holdingAssertion(t, me) {
		t.Fatalf("pmset shows no caffeinate assertion for pid %d", me)
	}

	// Active → toggle stops and resets to Off.
	toggle()
	if deps.Sup.Active() {
		t.Fatal("Sup.Active = true after toggle-off")
	}
	if selected != "Off" {
		t.Fatalf("selected = %q after toggle-off, want Off", selected)
	}
	if holdingAssertion(t, me) {
		t.Fatalf("pmset still shows caffeinate assertion for pid %d after toggle-off", me)
	}

	// Persisted 15m → toggle restores it.
	if err := config.Save(config.Config{PresetSecs: 900}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	toggle()
	if !deps.Sup.Active() {
		t.Fatal("Sup.Active = false after toggle restore")
	}
	if selected != "15m" {
		t.Fatalf("selected = %q after toggle restore, want 15m", selected)
	}
	toggle() // Release the hold.
	if deps.Sup.Active() {
		t.Fatal("Sup.Active = true after final toggle-off")
	}
}

// holdingAssertion reports whether pmset lists a caffeinate assertion owned
// by pid (caffeinate is spawned with -w pid, and pmset prints pid names).
func holdingAssertion(t *testing.T, pid int) bool {
	t.Helper()
	out, err := exec.Command("pmset", "-g", "assertions").Output()
	if err != nil {
		t.Fatalf("pmset: %v", err)
	}
	for line := range strings.Lines(string(out)) {
		if strings.Contains(line, "caffeinate") && strings.Contains(line, strconv.Itoa(pid)) {
			return true
		}
	}
	return false
}

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathOverride(t *testing.T) {
	t.Setenv("KOPI_O_CONFIG", "/tmp/ko-test-config.json")
	if got := Path(); got != "/tmp/ko-test-config.json" {
		t.Fatalf("Path() = %q, want override", got)
	}
}

func TestPathXDG(t *testing.T) {
	t.Setenv("KOPI_O_CONFIG", "")
	t.Setenv("XDG_CONFIG_HOME", "/tmp/ko-xdg")
	want := filepath.Join("/tmp/ko-xdg", "kopi-o", "config.json")
	if got := Path(); got != want {
		t.Fatalf("Path() = %q, want %q", got, want)
	}
}

func TestLoadMissingReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	got := LoadFrom(filepath.Join(dir, "nope", "config.json"))
	if got != Default() {
		t.Fatalf("LoadFrom(missing) = %+v, want %+v", got, Default())
	}
	if got.PresetSecs != 3600 {
		t.Fatalf("default preset = %d, want 3600", got.PresetSecs)
	}
}

func TestLoadCorruptReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")
	if err := os.WriteFile(p, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := LoadFrom(p); got != Default() {
		t.Fatalf("LoadFrom(corrupt) = %+v, want %+v", got, Default())
	}
}

func TestSaveRoundTripAtomicPerms(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "sub", "config.json")
	t.Setenv("KOPI_O_CONFIG", p)

	want := Config{PresetSecs: 900, Indefinite: false, LaunchAtLogin: true}
	if err := Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if got := Load(); got != want {
		t.Fatalf("Load() = %+v, want %+v", got, want)
	}

	fi, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if perm := fi.Mode().Perm(); perm != 0o600 {
		t.Fatalf("file perm = %o, want 600", perm)
	}
	di, err := os.Stat(filepath.Dir(p))
	if err != nil {
		t.Fatal(err)
	}
	if perm := di.Mode().Perm(); perm != 0o700 {
		t.Fatalf("dir perm = %o, want 700", perm)
	}
	// No tmp leftovers.
	entries, err := os.ReadDir(filepath.Dir(p))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if len(e.Name()) > 8 && e.Name()[:8] == ".config-" {
			t.Fatalf("leftover tmp file %q", e.Name())
		}
	}
}

func TestSaveIndefiniteRoundTrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")
	want := Config{Indefinite: true, LaunchAtLogin: false}
	if err := SaveTo(p, want); err != nil {
		t.Fatal(err)
	}
	if got := LoadFrom(p); got != want {
		t.Fatalf("LoadFrom() = %+v, want %+v", got, want)
	}
}

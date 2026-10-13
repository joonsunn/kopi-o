package tray

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// isolate HOME so tests never touch the real ~/Library.
func tempHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	return home
}

func TestLaunchAgentRoundTrip(t *testing.T) {
	home := tempHome(t)
	const exe = "/Users/test/Applications/Kopi-O.app/Contents/MacOS/kopi-o"

	if IsLaunchAtLogin() {
		t.Fatal("IsLaunchAtLogin = true before install")
	}
	if err := SetLaunchAtLogin(true, exe); err != nil {
		t.Fatalf("SetLaunchAtLogin(true): %v", err)
	}
	if !IsLaunchAtLogin() {
		t.Fatal("IsLaunchAtLogin = false after install")
	}
	data, err := os.ReadFile(LaunchAgentPath())
	if err != nil {
		t.Fatal(err)
	}
	wantPath := filepath.Join(home, "Library", "LaunchAgents", LaunchAgentLabel+".plist")
	if got := LaunchAgentPath(); got != wantPath {
		t.Fatalf("LaunchAgentPath() = %q, want %q", got, wantPath)
	}
	for _, want := range []string{LaunchAgentLabel, exe, "RunAtLoad"} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("plist missing %q:\n%s", want, data)
		}
	}

	if err := SetLaunchAtLogin(false, ""); err != nil {
		t.Fatalf("SetLaunchAtLogin(false): %v", err)
	}
	if IsLaunchAtLogin() {
		t.Fatal("IsLaunchAtLogin = true after removal")
	}
	// Removing twice is a no-op.
	if err := SetLaunchAtLogin(false, ""); err != nil {
		t.Fatalf("second remove: %v", err)
	}
}

func TestLaunchPlistShape(t *testing.T) {
	out := launchPlist("/Applications/Kopi-O.app/Contents/MacOS/kopi-o")
	for _, want := range []string{
		`<?xml version="1.0"`,
		`<key>Label</key>`,
		`<key>ProgramArguments</key>`,
		`<key>RunAtLoad</key>`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("plist missing %q:\n%s", want, out)
		}
	}
}

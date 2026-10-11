package tray

import (
	"testing"
	"time"
)

func TestTitle(t *testing.T) {
	cases := []struct {
		name       string
		active     bool
		indefinite bool
		rem        time.Duration
		want       string
	}{
		{"idle", false, false, 0, "○"},
		{"idle ignores remaining", false, false, time.Hour, "○"},
		{"indefinite", true, true, 0, "☕ ∞"},
		{"seconds", true, false, 45 * time.Second, "☕ 45s"},
		{"minutes", true, false, 42 * time.Minute, "☕ 42m"},
		{"minutes rounds down", true, false, 42*time.Minute + 50*time.Second, "☕ 42m"},
		{"hours", true, false, 2*time.Hour + 5*time.Minute, "☕ 2h05m"},
		{"negative clamps", true, false, -time.Second, "☕ 0s"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Title(c.active, c.indefinite, c.rem); got != c.want {
				t.Fatalf("Title() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestLabelFor(t *testing.T) {
	if got := labelFor(3600, false); got != "1h" {
		t.Fatalf("labelFor(3600) = %q, want 1h", got)
	}
	if got := labelFor(0, true); got != "∞" {
		t.Fatalf("labelFor(indef) = %q, want ∞", got)
	}
	if got := labelFor(60, false); got != "" {
		t.Fatalf("labelFor(60) = %q, want empty (custom)", got)
	}
}

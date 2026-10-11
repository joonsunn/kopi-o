package tray

import (
	"fmt"
	"time"
)

// Title renders the menu-bar title for the current hold state.
// The tray icon carries on/off state, so the title is countdown-only:
// idle is "" (icon only), indefinite is "∞", finite is like "42m".
func Title(active, indefinite bool, remaining time.Duration) string {
	if !active {
		return ""
	}
	if indefinite {
		return "∞"
	}
	if remaining < 0 {
		remaining = 0
	}
	switch {
	case remaining < 90*time.Second:
		return fmt.Sprintf("%ds", int(remaining.Seconds()))
	case remaining < 90*time.Minute:
		return fmt.Sprintf("%dm", int(remaining.Minutes()))
	default:
		h := int(remaining.Hours())
		m := int(remaining.Minutes()) % 60
		return fmt.Sprintf("%dh%02dm", h, m)
	}
}

// Tooltip renders the hover text for the current hold state, teaching
// click semantics. lastPreset is the label a click would start when idle
// (ignored when active).
func Tooltip(active, indefinite bool, remaining time.Duration, lastPreset string) string {
	if !active {
		return fmt.Sprintf("kopi-o: idle — click for %s, right-click for menu", lastPreset)
	}
	if indefinite {
		return "kopi-o: holding ∞ — click to stop, right-click for menu"
	}
	return fmt.Sprintf("kopi-o: holding %s — click to stop, right-click for menu",
		remaining.Truncate(time.Second).String())
}

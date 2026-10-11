package tray

import (
	"fmt"
	"time"
)

// Title renders the menu-bar title for the current hold state.
// Idle: "○". Indefinite: "☕ ∞". Finite: countdown like "☕ 42m".
func Title(active, indefinite bool, remaining time.Duration) string {
	if !active {
		return "○"
	}
	if indefinite {
		return "☕ ∞"
	}
	if remaining < 0 {
		remaining = 0
	}
	switch {
	case remaining < 90*time.Second:
		return fmt.Sprintf("☕ %ds", int(remaining.Seconds()))
	case remaining < 90*time.Minute:
		return fmt.Sprintf("☕ %dm", int(remaining.Minutes()))
	default:
		h := int(remaining.Hours())
		m := int(remaining.Minutes()) % 60
		return fmt.Sprintf("☕ %dh%02dm", h, m)
	}
}

// Tooltip renders the hover text for the current hold state.
func Tooltip(active, indefinite bool, remaining time.Duration) string {
	if !active {
		return "kopi-o: idle (sleep allowed)"
	}
	if indefinite {
		return "kopi-o: holding sleep indefinitely"
	}
	return "kopi-o: holding sleep, " + remaining.Truncate(time.Second).String() + " left"
}

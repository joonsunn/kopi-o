// Package sleep supervises a single /usr/bin/caffeinate child process.
//
// Start kills any prior holder before spawning the new one; Stop kills.
// A finite hold arms an expiry Timer that auto-stops and fires onExpire.
// Indefinite holds pass no -t flag. Mutex-guarded single holder.
package sleep

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

// DefaultPath is the caffeinate binary.
const DefaultPath = "/usr/bin/caffeinate"

// Supervisor holds at most one caffeinate process.
type Supervisor struct {
	mu       sync.Mutex
	path     string
	cmd      *exec.Cmd
	timer    *time.Timer
	active   bool
	secs     int
	infinite bool
	deadline time.Time
	onExpire func()
}

// New returns a Supervisor using DefaultPath. onExpire may be nil;
// it is invoked (once per expiry) after a finite hold expires.
func New(onExpire func()) *Supervisor {
	return &Supervisor{path: DefaultPath, onExpire: onExpire}
}

// NewWithPath is a test hook allowing a fake binary path.
func NewWithPath(path string, onExpire func()) *Supervisor {
	return &Supervisor{path: path, onExpire: onExpire}
}

// Args reports the last-spawned caffeinate args (test hook).
func (s *Supervisor) Args() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cmd == nil {
		return nil
	}
	return append([]string(nil), s.cmd.Args[1:]...)
}

// Start begins a hold: indefinite (no -t) or finite (-t secs).
// It kills any prior holder first. Finite secs must be > 0.
// The child watches the parent pid (-w) so a crash releases the hold.
func (s *Supervisor) Start(secs int, indefinite bool) error {
	if !indefinite && secs <= 0 {
		return errors.New("sleep: finite hold needs secs > 0")
	}
	args := []string{"-dimsu", "-w", strconv.Itoa(os.Getpid())}
	if !indefinite {
		args = append(args, "-t", strconv.Itoa(secs))
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopLocked()

	cmd := exec.Command(s.path, args...)
	// Detach output; caffeinate is quiet anyway.
	if err := cmd.Start(); err != nil {
		return err
	}
	// Reap in background so zombies don't accumulate.
	go cmd.Wait() //nolint:errcheck

	s.cmd = cmd
	s.active = true
	s.secs = secs
	s.infinite = indefinite
	if indefinite {
		s.deadline = time.Time{}
	} else {
		s.deadline = time.Now().Add(time.Duration(secs) * time.Second)
		d := time.Duration(secs) * time.Second
		s.timer = time.AfterFunc(d, s.expire)
	}
	return nil
}

// Stop ends any active hold. Safe to call when idle.
func (s *Supervisor) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopLocked()
}

// Active reports whether a hold is in effect.
func (s *Supervisor) Active() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}

// Remaining reports time left on a finite hold, or 0 when idle/indefinite.
func (s *Supervisor) Remaining() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.active || s.infinite {
		return 0
	}
	return time.Until(s.deadline)
}

// Indefinite reports whether the active hold has no timeout.
func (s *Supervisor) Indefinite() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active && s.infinite
}

// stopLocked kills the child and disarms the timer. Caller holds s.mu.
func (s *Supervisor) stopLocked() {
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}
	s.cmd = nil
	s.active = false
	s.infinite = false
	s.secs = 0
	s.deadline = time.Time{}
}

// expire runs on the Timer goroutine: clear state, then fire callback.
func (s *Supervisor) expire() {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	cb := s.onExpire
	s.stopLocked()
	s.mu.Unlock()
	if cb != nil {
		cb()
	}
}

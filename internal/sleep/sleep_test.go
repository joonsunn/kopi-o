package sleep

import (
	"slices"
	"syscall"
	"testing"
	"time"
)

func TestStartStop(t *testing.T) {
	s := New(nil)
	if err := s.Start(60, false); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !s.Active() {
		t.Fatal("Active = false after Start")
	}
	args := s.Args()
	if !slices.Contains(args, "-t") {
		t.Fatalf("finite hold args %q missing -t", args)
	}
	for _, want := range []string{"-dimsu", "-w"} {
		if !slices.Contains(args, want) {
			t.Fatalf("args %q missing %q", args, want)
		}
	}
	s.Stop()
	if s.Active() {
		t.Fatal("Active = true after Stop")
	}
	// Second stop is a safe no-op.
	s.Stop()
}

func TestRestartKillsPrior(t *testing.T) {
	s := New(nil)
	if err := s.Start(60, false); err != nil {
		t.Fatal(err)
	}
	s.mu.Lock()
	old := s.cmd
	s.mu.Unlock()
	if old == nil || old.Process == nil {
		t.Fatal("no process after Start")
	}
	oldPid := old.Process.Pid

	if err := s.Start(60, false); err != nil {
		t.Fatal(err)
	}
	s.mu.Lock()
	newCmd := s.cmd
	s.mu.Unlock()
	if newCmd == nil || newCmd.Process == nil {
		t.Fatal("no process after restart")
	}
	if newCmd.Process.Pid == oldPid {
		t.Fatal("restart reused pid, expected new process")
	}
	// Old process must be gone: Signal(0) on a dead pid errors.
	// (Small grace for Kill to land.)
	deadline := time.Now().Add(2 * time.Second)
	for {
		if err := old.Process.Signal(syscall.Signal(0)); err != nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("old pid %d still alive after restart", oldPid)
		}
		time.Sleep(20 * time.Millisecond)
	}
	s.Stop()
}

func TestExpiryFiresCallback(t *testing.T) {
	fired := make(chan struct{}, 1)
	s := New(func() { fired <- struct{}{} })
	if err := s.Start(1, false); err != nil {
		t.Fatal(err)
	}
	select {
	case <-fired:
	case <-time.After(5 * time.Second):
		t.Fatal("onExpire not fired after -t 1 hold")
	}
	if s.Active() {
		t.Fatal("Active = true after expiry")
	}
}

func TestIndefiniteHasNoTimeoutFlag(t *testing.T) {
	s := New(nil)
	if err := s.Start(0, true); err != nil {
		t.Fatal(err)
	}
	defer s.Stop()
	if !s.Active() || !s.Indefinite() {
		t.Fatal("expected active indefinite hold")
	}
	if args := s.Args(); slices.Contains(args, "-t") {
		t.Fatalf("indefinite args %q must not contain -t", args)
	}
	if rem := s.Remaining(); rem != 0 {
		t.Fatalf("Remaining() = %v for indefinite, want 0", rem)
	}
}

func TestInvalidSecs(t *testing.T) {
	s := New(nil)
	if err := s.Start(0, false); err == nil {
		t.Fatal("expected error for secs=0 finite hold")
		s.Stop()
	}
	if err := s.Start(-5, false); err == nil {
		t.Fatal("expected error for negative secs")
		s.Stop()
	}
	if s.Active() {
		t.Fatal("Active after failed Start")
	}
}

func TestBadBinary(t *testing.T) {
	s := NewWithPath("/nonexistent/caffeinate-xyz", nil)
	if err := s.Start(60, false); err == nil {
		t.Fatal("expected error for bad binary path")
		s.Stop()
	}
}

package server

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

type fakePurger struct {
	mu     sync.Mutex
	calls  int
	cutoff time.Time
	err    error
}

func (f *fakePurger) PurgeExpiredSessions(_ context.Context, cutoff time.Time) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.cutoff = cutoff
	return int64(f.calls), f.err
}

func (f *fakePurger) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatal("condition not met within 2s")
}

func TestRunSessionPurgeLoopPurgesAndStopsOnCancel(t *testing.T) {
	purger := &fakePurger{}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		defer close(done)
		RunSessionPurgeLoop(ctx, purger, 10*time.Millisecond, zerolog.Nop())
	}()

	// Первый purge выполняется сразу при старте.
	waitFor(t, func() bool { return purger.callCount() >= 1 })

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("purge loop did not stop after ctx cancel")
	}
}

func TestRunSessionPurgeLoopKeepsTickingOnError(t *testing.T) {
	purger := &fakePurger{err: context.DeadlineExceeded}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})

	go func() {
		defer close(done)
		RunSessionPurgeLoop(ctx, purger, 5*time.Millisecond, zerolog.Nop())
	}()

	// Ошибка очистки не должна останавливать цикл: ждём второго тика.
	waitFor(t, func() bool { return purger.callCount() >= 2 })

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("purge loop did not stop after ctx cancel")
	}
}

func TestRunSessionPurgeLoopPassesRetentionCutoff(t *testing.T) {
	purger := &fakePurger{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	started := time.Now()
	go RunSessionPurgeLoop(ctx, purger, time.Hour, zerolog.Nop())

	waitFor(t, func() bool { return purger.callCount() >= 1 })

	purger.mu.Lock()
	cutoff := purger.cutoff
	purger.mu.Unlock()

	retention := started.Sub(cutoff)
	if retention < sessionRetention-time.Minute || retention > sessionRetention+time.Minute {
		t.Fatalf("cutoff retention = %s, want ~%s", retention, sessionRetention)
	}
}

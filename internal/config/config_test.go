package config

import (
	"testing"
	"time"
)

func TestSessionsPurgeIntervalDefault(t *testing.T) {
	cfg := Config{}
	if got := cfg.SessionsPurgeInterval(); got != time.Hour {
		t.Fatalf("default interval: got %s, want 1h", got)
	}
}

func TestSessionsPurgeIntervalFromMinutes(t *testing.T) {
	cfg := Config{SessionsPurgeIntervalMinutes: 15}
	if got := cfg.SessionsPurgeInterval(); got != 15*time.Minute {
		t.Fatalf("configured interval: got %s, want 15m", got)
	}
}

func TestSessionsPurgeIntervalNegativeFallsBackToDefault(t *testing.T) {
	cfg := Config{SessionsPurgeIntervalMinutes: -5}
	if got := cfg.SessionsPurgeInterval(); got != time.Hour {
		t.Fatalf("negative interval: got %s, want 1h", got)
	}
}

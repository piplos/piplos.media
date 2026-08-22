package server

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

// SessionPurger removes stale refresh session rows.
type SessionPurger interface {
	PurgeExpiredSessions(ctx context.Context, cutoff time.Time) (int64, error)
}

// sessionRetention keeps revoked sessions auditable for as long as a refresh
// chain could still be live (matches the default refresh TTL).
const sessionRetention = 7 * 24 * time.Hour

// RunSessionPurgeLoop deletes expired and revoked sessions once immediately,
// then on every tick of interval, until ctx is cancelled. Intended to run in
// its own goroutine; it returns when the context is done.
func RunSessionPurgeLoop(ctx context.Context, purger SessionPurger, interval time.Duration, log zerolog.Logger) {
	if interval <= 0 {
		interval = time.Hour
	}

	purge := func() {
		removed, err := purger.PurgeExpiredSessions(ctx, time.Now().Add(-sessionRetention))
		switch {
		case err != nil:
			log.Error().Err(err).Msg("purge expired sessions")
		case removed > 0:
			log.Debug().Int64("removed", removed).Msg("purged expired sessions")
		}
	}

	if ctx.Err() != nil {
		return
	}
	purge()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			purge()
		}
	}
}

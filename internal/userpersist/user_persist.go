// Package userpersist owns the lazy-save wiring for online users.
//
// Any logic that mutates persistent fields on a game.OnlineUser (HP,
// future stats, etc.) should call SaveOrUpdate here so the change is
// eventually flushed to the database. Multiple calls within the
// lazy-save quiet window collapse into a single SQL UPDATE.
package userpersist

import (
	"strconv"

	"herostory-server/internal/game"
	"herostory-server/internal/repository"
	asyncop "herostory-server/pkg/async_op"
	lazysave "herostory-server/pkg/lazy_save"

	"github.com/rs/zerolog/log"
)

// lsoIDPrefix namespaces every user entry so other future lazy-save
// subjects cannot collide.
const lsoIDPrefix = "user:"

// lsoID builds the lazy-save key for a user id. Using strconv is much
// cheaper than fmt.Sprintf on this hot path.
func lsoID(userID int) string {
	return lsoIDPrefix + strconv.Itoa(userID)
}

// persistHP returns a persist closure that, when invoked by the
// lazy-save flusher, dispatches the DB write onto the worker pool
// keyed by userID. Keying preserves per-user FIFO ordering and keeps
// one user's slow query from blocking another's.
//
// The HP value is captured by value at closure creation time, so a
// later mutation to the OnlineUser does not corrupt an already-queued
// persist (though the next SaveOrUpdate will overwrite the queued
// closure before it gets a chance to run).
func persistHP(userID int, currHP int32) func() {
	return func() {
		asyncop.Process(userID, func() {
			if err := repository.UpdateCurrHp(userID, currHP); err != nil {
				log.Error().
					Err(err).
					Int("userId", userID).
					Int32("currHp", currHP).
					Msg("persist curr_hp failed")
			}
		}, nil)
	}
}

// SaveOrUpdate registers a delayed DB write for u's mutable fields.
// Repeated calls for the same user within the lazy-save quiet window
// collapse into a single UPDATE that carries the most recent HP.
func SaveOrUpdate(u *game.OnlineUser) {
	if u == nil || u.UserID <= 0 {
		return
	}
	lazysave.SaveOrUpdate(lsoID(u.UserID), persistHP(u.UserID, u.CurrHp))
}

// PersistNow synchronously writes u's mutable fields to the database on
// the calling goroutine. Unlike SaveOrUpdate it bypasses both the
// lazy-save quiet period and the asyncop worker pool, so by the time
// PersistNow returns the UPDATE has either succeeded or failed (and
// been logged).
//
// This is intentionally synchronous: the disconnect path calls it right
// before forgetting the user, and a fire-and-forget dispatch there
// could lose the write if the process is restarted, the worker pool's
// task is preempted, or the DB connection is closed before the queued
// task drains. Disconnect is not a hot path, so a single blocking
// UPDATE is acceptable; in exchange we get the "Now" semantics the
// name promises.
//
// After the synchronous write we drop any lazy-save record still
// registered for this user. Otherwise the flusher could later replay
// a stale snapshot (e.g. an HP=90 captured just before disconnect)
// on top of a newer authoritative value (e.g. HP=100 written by the
// next login's respawn repair), silently rewinding the user's HP.
func PersistNow(u *game.OnlineUser) {
	if u == nil || u.UserID <= 0 {
		return
	}
	if err := repository.UpdateCurrHp(u.UserID, u.CurrHp); err != nil {
		log.Error().
			Err(err).
			Int("userId", u.UserID).
			Int32("currHp", u.CurrHp).
			Msg("PersistNow: persist curr_hp failed")
	}
	lazysave.Cancel(lsoID(u.UserID))
}

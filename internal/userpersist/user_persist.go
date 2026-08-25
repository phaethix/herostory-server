// Package userpersist is the only place that should register online-user
// HP writes with lazy_save, so coalescing and cancel-on-disconnect stay
// consistent.
package userpersist

import (
	"strconv"

	"herostory-server/internal/game"
	"herostory-server/internal/repository"
	asyncop "herostory-server/pkg/async_op"
	lazysave "herostory-server/pkg/lazy_save"

	"github.com/rs/zerolog/log"
)

const lsoIDPrefix = "user:"

func lsoID(userID int) string {
	return lsoIDPrefix + strconv.Itoa(userID)
}

// persistHP captures currHP by value so a later in-memory mutation
// cannot change a snapshot already handed to the flusher.
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

// SaveOrUpdate coalesces this user's HP write into one SQL UPDATE after the quiet period.
func SaveOrUpdate(u *game.OnlineUser) {
	if u == nil || u.UserID <= 0 {
		return
	}
	lazysave.SaveOrUpdate(lsoID(u.UserID), persistHP(u.UserID, u.CurrHp))
}

// PersistNow blocks because disconnect is about to forget the user;
// a queued write could be lost on restart. Cancel afterwards so the
// flusher cannot replay an older snapshot over the value we just wrote.
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

// Package hero contains the select-hero game logic.
//
// Callers pass a command and receive a SelectHeroResult to write back to
// the requesting client. An empty HeroAvatar on the result means failure
// (see the proto comment on SelectHeroResult).
package hero

import (
	"herostory-server/internal/game"
	"herostory-server/internal/pb"
	"herostory-server/internal/repository"
	asyncop "herostory-server/pkg/async_op"

	"github.com/rs/zerolog/log"
)

// Apply assigns heroAvatar to the online user and returns the result to
// send back to that client. It returns nil when the command must be
// ignored (offline user or nil cmd). An empty avatar is a client-visible
// failure: the result is non-nil with HeroAvatar="" and in-memory state
// is left unchanged.
func Apply(userID int, cmd *pb.SelectHeroCmd) *pb.SelectHeroResult {
	if cmd == nil || userID <= 0 {
		return nil
	}

	user := game.GetOnlineUser(userID)
	if user == nil {
		log.Warn().
			Int("userId", userID).
			Msg("select hero ignored: not online")
		return nil
	}

	avatar := cmd.GetHeroAvatar()
	if avatar == "" {
		return &pb.SelectHeroResult{}
	}

	user.HeroAvatar = avatar
	persistAvatar(userID, avatar)

	return &pb.SelectHeroResult{HeroAvatar: avatar}
}

func persistAvatar(userID int, avatar string) {
	asyncop.Process(userID, func() {
		if err := repository.UpdateHeroAvatar(userID, avatar); err != nil {
			log.Error().
				Err(err).
				Int("userId", userID).
				Str("heroAvatar", avatar).
				Msg("persist hero_avatar failed")
		}
	}, nil)
}

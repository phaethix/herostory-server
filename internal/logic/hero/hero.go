// Package hero owns select-hero. An empty HeroAvatar on the result is
// a client-visible failure (see SelectHeroResult in the proto); nil
// means the command was dropped.
package hero

import (
	"herostory-server/internal/game"
	"herostory-server/internal/pb"
	"herostory-server/internal/repository"
	asyncop "herostory-server/pkg/async_op"

	"github.com/rs/zerolog/log"
)

func Apply(userID int, cmd *pb.SelectHeroCmd) *pb.SelectHeroResult {
	if cmd == nil || userID <= 0 {
		return nil
	}

	user := game.GetOnlineUser(userID)
	if user == nil {
		log.Warn().Int("userId", userID).Msg("select hero ignored: not online")
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

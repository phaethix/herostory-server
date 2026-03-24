package handler

import (
	"herostory-server/internal/game"
	"herostory-server/internal/network/broadcaster"
	"herostory-server/internal/pb"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/dynamicpb"
)

func init() {
	cmdHandlerMap[uint16(pb.MsgCode_USER_ENTRY_CMD)] = userEntryCmdHandler
}

func userEntryCmdHandler(ctx CmdContext, _ *dynamicpb.Message) {
	if ctx == nil || ctx.GetUserId() <= 0 {
		return
	}

	userId := int(ctx.GetUserId())
	user := game.GetOnlineUser(userId)
	if user == nil {
		log.Warn().
			Int("userId", userId).
			Msg("user entry ignored: user not in online list")
		return
	}

	rest := &pb.UserEntryResult{
		UserId:     uint32(userId),
		UserName:   user.UserName,
		HeroAvatar: user.HeroAvatar,
	}

	// broadcast user entry to all connected clients
	broadcaster.Broadcast(rest)
}

package handler

import (
	"context"

	"herostory-server/internal/logic/attack"
	"herostory-server/internal/network/broadcaster"
	"herostory-server/internal/pb"
	"herostory-server/internal/rank"
	asyncop "herostory-server/pkg/async_op"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/dynamicpb"
)

func init() {
	cmdHandlerMap[uint16(pb.MsgCode_USER_ATTK_CMD)] = userAttkCmdHandler
}

func userAttkCmdHandler(ctx CmdContext, msg *dynamicpb.Message) {
	if ctx == nil || msg == nil || ctx.GetUserId() <= 0 {
		return
	}

	cmd := unmarshalCmd[pb.UserAttkCmd](msg)

	result := attack.Apply(int(ctx.GetUserId()), cmd)
	if result == nil {
		return
	}

	for _, m := range result.Broadcasts {
		broadcaster.Broadcast(m)
	}
	if result.LoserID == 0 {
		return
	}

	winner, loser := result.WinnerID, result.LoserID
	asyncop.Process(winner, func() {
		if err := rank.RecordKill(context.Background(), winner, loser); err != nil {
			log.Error().
				Err(err).
				Int("winnerId", winner).
				Int("loserId", loser).
				Msg("rank: record kill failed")
		}
	}, nil)
}

package handler

import (
	"context"

	"herostory-server/internal/pb"
	"herostory-server/internal/rank"
	asyncop "herostory-server/pkg/async_op"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/dynamicpb"
)

func init() {
	cmdHandlerMap[uint16(pb.MsgCode_GET_RANK_CMD)] = getRankCmdHandler
}

func getRankCmdHandler(ctx CmdContext, _ *dynamicpb.Message) {
	if ctx == nil || ctx.GetUserId() <= 0 {
		return
	}

	var items []rank.Item
	asyncop.Process(0, func() {
		got, err := rank.Top(context.Background(), 0)
		if err != nil {
			log.Error().Err(err).Msg("rank: top failed")
			return
		}
		items = got
	}, func() {
		ctx.WriteMsg(rankResult(items))
	})
}

func rankResult(items []rank.Item) *pb.GetRankResult {
	out := &pb.GetRankResult{
		RankItem: make([]*pb.GetRankResult_RankItem, len(items)),
	}
	for i, it := range items {
		out.RankItem[i] = &pb.GetRankResult_RankItem{
			RankId:     uint32(it.RankID),
			UserId:     uint32(it.UserID),
			UserName:   it.UserName,
			HeroAvatar: it.HeroAvatar,
			Win:        it.Win,
		}
	}
	return out
}

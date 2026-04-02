package handler

import (
	"herostory-server/internal/game"
	"herostory-server/internal/pb"

	"google.golang.org/protobuf/types/dynamicpb"
)

func init() {
	cmdHandlerMap[uint16(pb.MsgCode_WHO_ELSE_IS_HERE_CMD)] = whoElseIsHereCmdHandler
}

func whoElseIsHereCmdHandler(ctx CmdContext, _ *dynamicpb.Message) {
	if ctx == nil || ctx.GetUserId() <= 0 {
		return
	}

	result := &pb.WhoElseIsHereResult{}

	game.ForEachOnlineUser(func(u *game.OnlineUser) {
		info := &pb.WhoElseIsHereResult_UserInfo{
			UserId:     uint32(u.UserID),
			UserName:   u.UserName,
			HeroAvatar: u.HeroAvatar,
		}
		if u.MoveState != nil {
			info.MoveState = u.MoveState.ToPB()
		}
		result.UserInfo = append(result.UserInfo, info)
	})

	ctx.WriteMsg(result)
}

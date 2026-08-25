package handler

import (
	"herostory-server/internal/logic/hero"
	"herostory-server/internal/pb"

	"google.golang.org/protobuf/types/dynamicpb"
)

func init() {
	cmdHandlerMap[uint16(pb.MsgCode_SELECT_HERO_CMD)] = selectHeroCmdHandler
}

func selectHeroCmdHandler(ctx CmdContext, msg *dynamicpb.Message) {
	if ctx == nil || msg == nil || ctx.GetUserId() <= 0 {
		return
	}

	cmd := unmarshalCmd[pb.SelectHeroCmd](msg)

	result := hero.Apply(int(ctx.GetUserId()), cmd)
	if result == nil {
		return
	}

	// Reply to the requester only. Others see the new avatar on Entry / WhoElseIsHere.
	ctx.WriteMsg(result)
}

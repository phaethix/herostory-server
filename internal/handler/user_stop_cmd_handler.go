package handler

import (
	"herostory-server/internal/logic/move"
	"herostory-server/internal/network/broadcaster"
	"herostory-server/internal/pb"

	"google.golang.org/protobuf/types/dynamicpb"
)

func init() {
	cmdHandlerMap[uint16(pb.MsgCode_USER_STOP_CMD)] = userStopCmdHandler
}

func userStopCmdHandler(ctx CmdContext, _ *dynamicpb.Message) {
	if ctx == nil || ctx.GetUserId() <= 0 {
		return
	}

	result := move.Stop(int(ctx.GetUserId()))
	if result == nil {
		return
	}

	broadcaster.Broadcast(result)
}

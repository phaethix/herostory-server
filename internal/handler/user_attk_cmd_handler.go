package handler

import (
	"herostory-server/internal/logic/attack"
	"herostory-server/internal/network/broadcaster"
	"herostory-server/internal/pb"

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
}

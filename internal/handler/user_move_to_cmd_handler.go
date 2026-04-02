package handler

import (
	"herostory-server/internal/logic/move"
	"herostory-server/internal/network/broadcaster"
	"herostory-server/internal/pb"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func init() {
	cmdHandlerMap[uint16(pb.MsgCode_USER_MOVE_TO_CMD)] = userMoveToCmdHandler
}

func userMoveToCmdHandler(ctx CmdContext, msg *dynamicpb.Message) {
	if ctx == nil || msg == nil || ctx.GetUserId() <= 0 {
		return
	}

	cmd := &pb.UserMoveToCmd{}
	msg.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		cmd.ProtoReflect().Set(fd, v)
		return true
	})

	result := move.Apply(int(ctx.GetUserId()), cmd)
	if result == nil {
		return
	}

	broadcaster.Broadcast(result)
}

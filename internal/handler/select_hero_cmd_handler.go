package handler

import (
	"herostory-server/internal/logic/hero"
	"herostory-server/internal/pb"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func init() {
	cmdHandlerMap[uint16(pb.MsgCode_SELECT_HERO_CMD)] = selectHeroCmdHandler
}

func selectHeroCmdHandler(ctx CmdContext, msg *dynamicpb.Message) {
	if ctx == nil || msg == nil || ctx.GetUserId() <= 0 {
		return
	}

	cmd := &pb.SelectHeroCmd{}
	msg.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		cmd.ProtoReflect().Set(fd, v)
		return true
	})

	result := hero.Apply(int(ctx.GetUserId()), cmd)
	if result == nil {
		return
	}

	// SelectHeroResult is a reply to the requesting client, not a broadcast.
	// Other players pick up the new avatar via UserEntry / WhoElseIsHere.
	ctx.WriteMsg(result)
}

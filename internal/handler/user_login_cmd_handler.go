package handler

import (
	"herostory-server/internal/pb"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

const (
	HeroAvatarDefault = "Hero_Shaman"
)

func init() {
	cmdHandlerMap[uint16(pb.MsgCode_USER_LOGIN_CMD)] = userLoginCmdHandler
}

func userLoginCmdHandler(ctx CmdContext, msg *dynamicpb.Message) {
	if ctx == nil || msg == nil {
		return
	}

	cmd := &pb.UserLoginCmd{}
	msg.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		cmd.ProtoReflect().Set(fd, v)
		return true
	})

	rest := &pb.UserLoginResult{
		UserId:     1,
		UserName:   cmd.UserName,
		HeroAvatar: HeroAvatarDefault,
	}

	ctx.BindUserId(1)
	ctx.WriteMsg(rest)
}

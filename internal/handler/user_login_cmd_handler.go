package handler

import (
	"herostory-server/internal/logic/login"
	"herostory-server/internal/model"
	"herostory-server/internal/pb"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
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

	// login runs DB operations in a background goroutine and delivers
	// the result back to the main thread via callback.
	login.LoginByPasswordAsync(cmd.UserName, cmd.Password, func(user *model.User) {
		if user == nil {
			// login failed – userId 0 signals failure to the client
			ctx.WriteMsg(&pb.UserLoginResult{
				UserId:     0,
				UserName:   cmd.UserName,
				HeroAvatar: "",
			})
			return
		}

		// login successful – bind the user id to this connection
		ctx.BindUserId(int64(user.ID))
		ctx.WriteMsg(&pb.UserLoginResult{
			UserId:     uint32(user.ID),
			UserName:   user.UserName,
			HeroAvatar: user.HeroAvatar,
		})
	})
}

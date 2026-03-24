package handler

import (
	"herostory-server/internal/game"
	"herostory-server/internal/logic/login"
	"herostory-server/internal/pb"

	"github.com/rs/zerolog/log"
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

	// LoginByPasswordAsync returns a typed AsyncBizResult[model.User] immediately.
	// The actual DB I/O runs on an async worker goroutine.
	bizResult := login.LoginByPasswordAsync(cmd.UserName, cmd.Password)

	if bizResult == nil {
		log.Error().
			Str("username", cmd.UserName).
			Msg("biz result is nil")
		return
	}

	// OnComplete is dispatched to the main thread once the async operation
	// finishes and sets the returned object.
	bizResult.OnComplete(func() {
		user := bizResult.GetReturnedObj()

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

		// register user in the online user group (pure data, no connection)
		game.AddOnlineUser(&game.OnlineUser{
			UserID:     user.ID,
			UserName:   user.UserName,
			HeroAvatar: user.HeroAvatar,
		})

		ctx.WriteMsg(&pb.UserLoginResult{
			UserId:     uint32(user.ID),
			UserName:   user.UserName,
			HeroAvatar: user.HeroAvatar,
		})
	})
}

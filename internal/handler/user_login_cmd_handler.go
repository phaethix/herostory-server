package handler

import (
	"herostory-server/internal/game"
	"herostory-server/internal/logic/login"
	"herostory-server/internal/pb"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/dynamicpb"
)

func init() {
	cmdHandlerMap[uint16(pb.MsgCode_USER_LOGIN_CMD)] = userLoginCmdHandler
}

func userLoginCmdHandler(ctx CmdContext, msg *dynamicpb.Message) {
	if ctx == nil || msg == nil {
		return
	}

	cmd := unmarshalCmd[pb.UserLoginCmd](msg)
	bizResult := login.LoginByPasswordAsync(cmd.UserName, cmd.Password)
	if bizResult == nil {
		log.Error().
			Str("username", cmd.UserName).
			Msg("biz result is nil")
		return
	}

	bizResult.OnComplete(func() {
		user := bizResult.GetReturnedObj()
		if user == nil {
			// Wire contract: UserId 0 is the only failure signal the client understands.
			ctx.WriteMsg(&pb.UserLoginResult{
				UserId:     0,
				UserName:   cmd.UserName,
				HeroAvatar: "",
			})
			return
		}

		ctx.BindUserId(int64(user.ID))
		hp := login.SessionHP(user)

		game.AddOnlineUser(&game.OnlineUser{
			UserID:     user.ID,
			UserName:   user.UserName,
			HeroAvatar: user.HeroAvatar,
			CurrHp:     hp,
		})

		ctx.WriteMsg(&pb.UserLoginResult{
			UserId:     uint32(user.ID),
			UserName:   user.UserName,
			HeroAvatar: user.HeroAvatar,
		})
	})
}

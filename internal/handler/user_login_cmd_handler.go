package handler

import (
	"cmp"
	"herostory-server/internal/game"
	"herostory-server/internal/logic/login"
	"herostory-server/internal/model"
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

		// HP is the DB's source of truth: registration writes
		// DefaultMaxHp, attacks decrement it, and disconnect flushes
		// the latest value back. The cmp.Or fallback below covers two
		// edge cases:
		//   - legacy rows that pre-date the curr_hp column (or its
		//     default:100 change) and still hold 0;
		//   - any future case where a process crash skipped the final
		//     PersistNow on death (target.CurrHp == 0 in DB).
		// In both, treating 0 as "respawn at full HP" is the desired
		// gameplay behaviour. cmp.Or returns its first non-zero
		// argument; negative HP is not possible by design (attacks
		// only drive HP toward 0 and the column is NOT NULL).
		game.AddOnlineUser(&game.OnlineUser{
			UserID:     user.ID,
			UserName:   user.UserName,
			HeroAvatar: user.HeroAvatar,
			CurrHp:     cmp.Or(user.CurrHp, model.DefaultMaxHp),
		})

		ctx.WriteMsg(&pb.UserLoginResult{
			UserId:     uint32(user.ID),
			UserName:   user.UserName,
			HeroAvatar: user.HeroAvatar,
		})
	})
}

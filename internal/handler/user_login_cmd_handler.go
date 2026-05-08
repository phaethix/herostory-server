package handler

import (
	"herostory-server/internal/game"
	"herostory-server/internal/logic/login"
	"herostory-server/internal/model"
	"herostory-server/internal/pb"
	"herostory-server/internal/repository"

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

		// Login starts a fresh session at full HP, unconditionally.
		//
		// Rationale: the wire protocol (UserLoginResult /
		// UserEntryResult / WhoElseIsHereResult) carries no HP field,
		// so the client always renders a freshly-logged-in avatar at
		// full HP regardless of what the server thinks. If we kept the
		// DB's HP across sessions, a user who disconnected at e.g.
		// HP=10 would re-enter looking healthy on screen but die from
		// a single hit on the server — exactly the bug we hit in
		// practice. Until the protocol grows an HP field, the only
		// consistent contract is "login == respawn".
		//
		// curr_hp is therefore a session-scoped value: lazy-save still
		// protects against in-session crashes (so a hit landed seconds
		// before a server restart isn't lost mid-fight), but it does
		// not survive a clean logout/login cycle.
		hp := model.DefaultMaxHp
		if hp != user.CurrHp {
			if err := repository.UpdateCurrHp(user.ID, hp); err != nil {
				log.Error().
					Err(err).
					Int("userId", user.ID).
					Int32("currHp", hp).
					Msg("login: respawn HP repair failed")
			}
		}

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

package handler

import (
	"herostory-server/internal/codec"
	"herostory-server/internal/pb"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

const (
	HeroAvatarDefault = "Hero_Shaman"
)

func init() {
	cmdHandlerMap[uint16(pb.MsgCode_USER_LOGIN_CMD)] = userLoginCmdHandler
}

func userLoginCmdHandler(conn *websocket.Conn, msg *dynamicpb.Message) {
	if conn == nil || msg == nil {
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
	
	byteArray, err := codec.EncodeMessage(rest)
	if err != nil {
		log.Error().Msgf(
			"encode client %v login result failed, err: %v",
			conn.RemoteAddr(),
			err,
		)
		return
	}

	err = conn.WriteMessage(websocket.BinaryMessage, byteArray)
	if err != nil {
		log.Error().Msgf(
			"write client %v login result failed, err: %v",
			conn.RemoteAddr(),
			err,
		)
	}
}

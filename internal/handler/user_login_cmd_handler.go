package handler

import (
	"herostory-server/internal/pb"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func init() {
	cmdHandlerMap[uint16(pb.MsgCode_USER_LOGIN_CMD)] = UserLoginCmdHandler
}

func UserLoginCmdHandler(conn *websocket.Conn, obj *dynamicpb.Message) {
	if conn == nil || obj == nil {
		return
	}

	cmd := &pb.UserLoginCmd{}
	obj.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		cmd.ProtoReflect().Set(fd, v)
		return true
	})
}

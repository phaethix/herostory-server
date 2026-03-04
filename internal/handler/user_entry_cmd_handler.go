package handler

import (
	"herostory-server/internal/pb"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func init() {
	cmdHandlerMap[uint16(pb.MsgCode_USER_ENTRY_CMD)] = EntryCmdHandler
}

func EntryCmdHandler(conn *websocket.Conn, msg *dynamicpb.Message) {
	if conn == nil || msg == nil {
		return
	}

	cmd := &pb.UserEntryCmd{}
	msg.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		cmd.ProtoReflect().Set(fd, v)
		return true
	})
}

package handler

import "google.golang.org/protobuf/reflect/protoreflect"

// CmdContext abstracts the underlying transport (WebSocket, etc.)
// and provides methods to interact with the connected client.
type CmdContext interface {
	BindUserId(userId int64)
	GetUserId() int64
	GetClientAddr() string
	WriteMsg(msg protoreflect.ProtoMessage)
	SendErrorMsg(code int, msg string)
	Disconnect()
}

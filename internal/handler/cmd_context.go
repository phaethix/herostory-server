package handler

import "google.golang.org/protobuf/reflect/protoreflect"

type CmdContext interface {
	BindUserId(userId int64)
	GetUserId() int64
	GetClientAddr() string
	WriteMsg(msg protoreflect.ProtoMessage)
	SendErrorMsg(code int, msg string)
	Disconnect()
}

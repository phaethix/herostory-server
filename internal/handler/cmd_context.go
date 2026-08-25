package handler

import "google.golang.org/protobuf/reflect/protoreflect"

// CmdContext is the transport-facing handle a command handler may use.
// Implementations must be safe for WriteMsg from the main goroutine.
type CmdContext interface {
	BindUserId(userId int64)
	GetUserId() int64
	WriteMsg(msg protoreflect.ProtoMessage)
}

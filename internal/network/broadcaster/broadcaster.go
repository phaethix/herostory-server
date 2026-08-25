package broadcaster

import "google.golang.org/protobuf/reflect/protoreflect"

// MsgWriter is defined by this consumer so the broadcaster does not
// depend on the WebSocket package.
type MsgWriter interface {
	WriteMsg(msg protoreflect.ProtoMessage)
}

// innerMap is only touched from the main goroutine.
var innerMap = make(map[int32]MsgWriter)

func AddCmdCtx(sessionID int32, ctx MsgWriter) {
	if sessionID <= 0 || ctx == nil {
		return
	}
	innerMap[sessionID] = ctx
}

func RemoveCmdCtx(sessionID int32) {
	if sessionID <= 0 {
		return
	}
	delete(innerMap, sessionID)
}

// Broadcast writes msg to every registered session.
func Broadcast(msg protoreflect.ProtoMessage) {
	if msg == nil {
		return
	}
	for _, ctx := range innerMap {
		ctx.WriteMsg(msg)
	}
}

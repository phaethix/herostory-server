package broadcaster

import "google.golang.org/protobuf/reflect/protoreflect"

// MsgWriter is the minimal interface needed to send a message to a client.
// Defined here by the consumer (broadcaster) — Go convention.
type MsgWriter interface {
	WriteMsg(msg protoreflect.ProtoMessage)
}

// innerMap holds all active connections keyed by session ID.
// All methods are called from the main thread, so no lock is needed.
var innerMap = make(map[int32]MsgWriter)

// AddCmdCtx registers a connection context by session ID.
func AddCmdCtx(sessionID int32, ctx MsgWriter) {
	if sessionID <= 0 || ctx == nil {
		return
	}
	innerMap[sessionID] = ctx
}

// RemoveCmdCtx removes a connection context by session ID.
func RemoveCmdCtx(sessionID int32) {
	if sessionID <= 0 {
		return
	}
	delete(innerMap, sessionID)
}

// Broadcast sends a message to every connected client.
func Broadcast(msg protoreflect.ProtoMessage) {
	if msg == nil {
		return
	}
	for _, ctx := range innerMap {
		if ctx != nil {
			ctx.WriteMsg(msg)
		}
	}
}

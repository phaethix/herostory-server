package websocket

import (
	"encoding/binary"
	"herostory-server/internal/codec"
	"herostory-server/internal/game"
	"herostory-server/internal/handler"
	"herostory-server/internal/network/broadcaster"
	"herostory-server/internal/pb"
	"herostory-server/internal/userpersist"
	"herostory-server/pkg/main_thread"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	MsgQueueSize          = 1024
	OneSecond             = 1000
	ReadMsgCountPerSecond = 8
)

type CmdContext struct {
	userId    int64
	SessionID int32
	addr      string
	conn      *websocket.Conn
	msgQ      chan protoreflect.ProtoMessage
}

func NewCmdContext(conn *websocket.Conn, sessionID int32) *CmdContext {
	return &CmdContext{
		conn:      conn,
		SessionID: sessionID,
	}
}

func (w *CmdContext) BindUserId(userId int64) {
	w.userId = userId
}

func (w *CmdContext) GetUserId() int64 {
	return w.userId
}

func (w *CmdContext) GetClientAddr() string {
	return w.addr
}

func (w *CmdContext) WriteMsg(msg protoreflect.ProtoMessage) {
	if msg == nil || w.conn == nil || w.msgQ == nil {
		return
	}

	w.msgQ <- msg
}

func (w *CmdContext) SendErrorMsg(code int, msg string) {
	// Implementation for sending error message
}

func (w *CmdContext) Disconnect() {
	if w.conn != nil {
		_ = w.conn.Close()
	}
}

func (w *CmdContext) LoopSendMessage() {
	w.msgQ = make(chan protoreflect.ProtoMessage, MsgQueueSize)

	go func() {
		for msg := range w.msgQ {
			if msg == nil {
				continue
			}

			data, err := codec.EncodeMessage(msg)
			if err != nil {
				log.Error().
					Err(err).
					Str("client", w.conn.RemoteAddr().String()).
					Msg("encode message failed")
				continue
			}

			err = w.conn.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				log.Error().
					Err(err).
					Str("client", w.conn.RemoteAddr().String()).
					Msg("write message failed")
			}
		}
	}()
}

func (w *CmdContext) LoopReceiveMessage() {
	if w.conn == nil {
		return
	}

	defer w.cleanupOnDisconnect()

	w.conn.SetReadLimit(64 * 1024)

	t0, n := int64(0), 0

	for {
		_, data, err := w.conn.ReadMessage()
		if err != nil {
			log.Error().
				Err(err).
				Msg("read message failed")
			break
		}

		t1 := time.Now().UnixMilli()
		if t1-t0 > OneSecond {
			t0, n = t1, 0
		}

		if n > ReadMsgCountPerSecond {
			log.Warn().
				Str("client", w.conn.RemoteAddr().String()).
				Int("message_count", n).
				Msg("client is sending messages too fast")
			continue
		}
		n++

		code := binary.BigEndian.Uint16(data[2:4])
		msg, err := codec.DecodeMessage(data[4:], int16(code))
		if err != nil {
			log.Error().
				Uint16("code", code).
				Err(err).
				Msg("decode client message failed")
			continue
		}

		log.Info().
			Uint16("code", code).
			Str("message", string(msg.Descriptor().Name())).
			Msg("received client message")

		h := handler.CreateCmdHandler(code)
		if h == nil {
			log.Warn().
				Uint16("code", code).
				Msg("no handler found for client message")
			continue
		}

		main_thread.Process(func() { h(w, msg) })
	}
}

// cleanupOnDisconnect removes the user from the online list, force-flushes
// any pending mutable state (e.g. HP) to the DB, and broadcasts
// UserQuitResult to notify other players.
//
// This method is invoked from the connection's reader goroutine. All
// mutations to game-wide state are routed through main_thread.Process so
// they execute on the same goroutine as command handlers, preserving the
// "single-threaded game loop" invariant.
func (w *CmdContext) cleanupOnDisconnect() {
	uid := w.userId
	if uid <= 0 {
		return
	}

	main_thread.Process(func() {
		if u := game.GetOnlineUser(int(uid)); u != nil {
			// Persist the latest HP before forgetting the user. Using
			// PersistNow bypasses the lazy-save quiet period so a quick
			// log-in/log-out followed by a process restart does not
			// rewind the user's HP.
			userpersist.PersistNow(u)
		}

		game.RemoveOnlineUser(int(uid))

		broadcaster.Broadcast(&pb.UserQuitResult{
			QuitUserId: uint32(uid),
		})
	})
}
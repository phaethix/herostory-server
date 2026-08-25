package websocket

import (
	"encoding/binary"
	"time"

	"herostory-server/internal/codec"
	"herostory-server/internal/game"
	"herostory-server/internal/handler"
	"herostory-server/internal/network/broadcaster"
	"herostory-server/internal/pb"
	"herostory-server/internal/userpersist"
	"herostory-server/pkg/main_thread"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	MsgQueueSize          = 1024
	OneSecond             = 1000
	ReadMsgCountPerSecond = 8
	minFrameLen           = 4
)

type CmdContext struct {
	userId    int64
	SessionID int32
	addr      string
	conn      *websocket.Conn
	msgQ      chan protoreflect.ProtoMessage
}

// NewCmdContext creates the send queue immediately so Broadcast can
// WriteMsg during handshake, before LoopSendMessage has started.
func NewCmdContext(conn *websocket.Conn, sessionID int32) *CmdContext {
	ctx := &CmdContext{
		conn:      conn,
		SessionID: sessionID,
		msgQ:      make(chan protoreflect.ProtoMessage, MsgQueueSize),
	}
	if conn != nil {
		ctx.addr = conn.RemoteAddr().String()
	}
	return ctx
}

func (w *CmdContext) BindUserId(userId int64) { w.userId = userId }
func (w *CmdContext) GetUserId() int64        { return w.userId }

func (w *CmdContext) WriteMsg(msg protoreflect.ProtoMessage) {
	if msg == nil || w.msgQ == nil {
		return
	}
	w.msgQ <- msg
}

// LoopSendMessage starts the goroutine that writes queued protobufs to the socket.
func (w *CmdContext) LoopSendMessage() {
	go func() {
		for msg := range w.msgQ {
			if msg == nil {
				continue
			}
			data, err := codec.EncodeMessage(msg)
			if err != nil {
				log.Error().Err(err).Str("client", w.addr).Msg("encode message failed")
				continue
			}
			if err := w.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				log.Error().Err(err).Str("client", w.addr).Msg("write message failed")
			}
		}
	}()
}

// LoopReceiveMessage reads frames until the connection dies, then cleans up the session.
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
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Error().Err(err).Str("client", w.addr).Msg("read message failed")
			}
			break
		}

		t1 := time.Now().UnixMilli()
		if t1-t0 > OneSecond {
			t0, n = t1, 0
		}
		if n > ReadMsgCountPerSecond {
			log.Warn().
				Str("client", w.addr).
				Int("message_count", n).
				Msg("client is sending messages too fast")
			continue
		}
		n++

		if len(data) < minFrameLen {
			log.Warn().Str("client", w.addr).Int("len", len(data)).Msg("short websocket frame")
			continue
		}

		code := binary.BigEndian.Uint16(data[2:4])
		msg, err := codec.DecodeMessage(data[4:], int16(code))
		if err != nil {
			log.Error().Uint16("code", code).Err(err).Msg("decode client message failed")
			continue
		}

		log.Info().
			Uint16("code", code).
			Str("message", string(msg.Descriptor().Name())).
			Msg("received client message")

		h := handler.CreateCmdHandler(code)
		if h == nil {
			log.Warn().Uint16("code", code).Msg("no handler found for client message")
			continue
		}
		main_thread.Process(func() { h(w, msg) })
	}
}

func (w *CmdContext) cleanupOnDisconnect() {
	uid := w.userId
	if uid <= 0 {
		return
	}

	// Reader goroutine: hop onto the main thread so we don't race
	// command handlers on onlineUsers / broadcaster.
	main_thread.Process(func() {
		if u := game.GetOnlineUser(int(uid)); u != nil {
			userpersist.PersistNow(u)
		}
		game.RemoveOnlineUser(int(uid))
		broadcaster.Broadcast(&pb.UserQuitResult{QuitUserId: uint32(uid)})
	})
}

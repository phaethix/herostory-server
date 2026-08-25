package server

import (
	"net/http"
	"sync/atomic"

	"herostory-server/internal/network/broadcaster"
	wsctx "herostory-server/internal/network/websocket"
	"herostory-server/pkg/main_thread"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin:     func(*http.Request) bool { return true },
}

var sessionIDCounter atomic.Int32

// WebSocketHandshake upgrades the request and runs until the client disconnects.
func WebSocketHandshake(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("websocket upgrade failed")
		return
	}
	defer conn.Close()

	log.Info().
		Str("remote", conn.RemoteAddr().String()).
		Msg("client connected to websocket")

	sid := sessionIDCounter.Add(1)
	ctx := wsctx.NewCmdContext(conn, sid)

	// Drain the send queue before the session is visible to Broadcast,
	// otherwise WriteMsg can fill a channel nobody is reading.
	ctx.LoopSendMessage()

	main_thread.ProcessWait(func() { broadcaster.AddCmdCtx(sid, ctx) })
	defer main_thread.ProcessWait(func() { broadcaster.RemoveCmdCtx(sid) })

	ctx.LoopReceiveMessage()
}

package server

import (
	"herostory-server/internal/network/broadcaster"
	websocket2 "herostory-server/internal/network/websocket"
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// sessionIDCounter assigns a unique, monotonically increasing session
// ID to every accepted WebSocket connection.
var sessionIDCounter atomic.Int32

func WebSocketHandshake(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().
			Err(err).
			Msg("websocket upgrade failed")
		return
	}
	defer conn.Close()

	log.Info().
		Str("remote", conn.RemoteAddr().String()).
		Msg("client connected to websocket")

	sid := sessionIDCounter.Add(1)
	ctx := websocket2.NewCmdContext(conn, sid)

	broadcaster.AddCmdCtx(sid, ctx)
	defer broadcaster.RemoveCmdCtx(sid)

	ctx.LoopSendMessage()
	ctx.LoopReceiveMessage()
}

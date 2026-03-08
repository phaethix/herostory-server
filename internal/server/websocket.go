package server

import (
	websocket2 "herostory-server/internal/network/websocket"
	"net/http"

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

	ctx := websocket2.NewCmdContext(conn)

	ctx.LoopSendMessage()
	ctx.LoopReceiveMessage()
}

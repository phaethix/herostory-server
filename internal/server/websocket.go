package server

import (
	"encoding/binary"
	"errors"
	"herostory-server/internal/codec"
	"herostory-server/internal/handler"
	"herostory-server/internal/main_thread"
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

var (
	ErrUpgradeWebSocket = errors.New("websocket upgrade failed")
	ErrReadMessage      = errors.New("websocket read message failed")
)

func WebSocketHandshake(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Error().Msgf("%v: %v", ErrUpgradeWebSocket, err)
		return
	}
	defer conn.Close()

	log.Info().Msgf("client %v connected to websocket", conn.RemoteAddr())

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Error().Msgf("%v: %v", ErrReadMessage, err)
			break
		}

		code := binary.BigEndian.Uint16(data[2:4])
		msg, err := codec.DecodeMessage(data[4:], int16(code))
		if err != nil {
			log.Error().Msgf(
				"decode client %v message failed, code: %v, err: %v",
				conn.RemoteAddr(),
				code,
				err,
			)
			continue
		}

		log.Info().Msgf(
			"received client %v message => data: %v, code: %v, msg: %v",
			conn.RemoteAddr(),
			data,
			code,
			msg.Descriptor().Name(),
		)

		h := handler.CreateCmdHandler(code)
		if h == nil {
			log.Warn().Msgf(
				"no handler found for client %v message, code: %v",
				conn.RemoteAddr(),
				code,
			)
			continue
		}

		main_thread.Process(func() { h(conn, msg) })
	}
}

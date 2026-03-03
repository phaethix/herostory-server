package handler

import (
	"errors"
	"fmt"
	"herostory-server/internal/pb"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
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

		log.Info().Msgf("received client %v message: %v", conn.RemoteAddr(), data)

		// parse the login message sent by the client
		cmd := &pb.UserLoginCmd{}
		_ = proto.Unmarshal(data[4:], cmd)
		log.Info().Msgf("[cmd] unmarshal client %v message : %v", conn.RemoteAddr(), cmd)

		fmt.Printf("cmd.GetPassword(): %v\n", cmd.GetPassword())

		// fmt.Printf("cmd.ProtoReflect().Descriptor().Name(): %v\n", cmd.ProtoReflect().Descriptor().Name())
		// fmt.Printf("cmd.ProtoReflect().Descriptor().Fields(): %v\n", cmd.ProtoReflect().Descriptor().Fields())
		// pb.File_api_proto_game_msg_proto.Messages()

		
		desc := pb.File_api_proto_game_msg_proto.Messages().ByName("UserLoginCmd")
		var cmd2 *dynamicpb.Message = dynamicpb.NewMessage(desc)
		_ = proto.Unmarshal(data[4:], cmd2)
		log.Info().Msgf("[cmd2] unmarshal client %v message : %v", conn.RemoteAddr(), cmd2)
		
		fmt.Printf("cmd2.GetPassword(): %v\n", cmd2.Get(desc.Fields().ByName("password")))

		
		cmd2.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			cmd.ProtoReflect().Set(fd, v)
			return true
		})
		log.Info().Msgf("[cmd2->cmd] unmarshal client %v message : %v", conn.RemoteAddr(), cmd)

		fmt.Printf("[cmd2->cmd] cmd.GetPassword(): %v\n", cmd.GetPassword())
	}
}

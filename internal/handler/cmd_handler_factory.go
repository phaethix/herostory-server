package handler

import (
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/types/dynamicpb"
)

type CmdHandler func(conn *websocket.Conn, obj *dynamicpb.Message)

var cmdHandlerMap = make(map[uint16]CmdHandler)

func CreateCmdHandler(code uint16) CmdHandler { return cmdHandlerMap[code] }

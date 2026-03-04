package handler

import (
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/types/dynamicpb"
)

type cmdHandlerFunc func(conn *websocket.Conn, obj *dynamicpb.Message)

var cmdHandlerMap = make(map[uint16]cmdHandlerFunc)

func CreateCmdHandler(code uint16) cmdHandlerFunc { return cmdHandlerMap[code] }

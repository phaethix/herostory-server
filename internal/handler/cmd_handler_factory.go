package handler

import (
	"google.golang.org/protobuf/types/dynamicpb"
)

type CmdHandlerFunc func(ctx CmdContext, obj *dynamicpb.Message)

var cmdHandlerMap = make(map[uint16]CmdHandlerFunc)

func CreateCmdHandler(code uint16) CmdHandlerFunc { return cmdHandlerMap[code] }

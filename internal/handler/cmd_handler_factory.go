package handler

import (
	"google.golang.org/protobuf/types/dynamicpb"
)

type CmdHandlerFunc func(ctx CmdContext, obj *dynamicpb.Message)

var cmdHandlerMap = make(map[uint16]CmdHandlerFunc)

// CreateCmdHandler returns the handler for code, or nil if none is registered.
func CreateCmdHandler(code uint16) CmdHandlerFunc { return cmdHandlerMap[code] }

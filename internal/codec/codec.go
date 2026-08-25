package codec

import (
	"encoding/binary"
	"slices"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// DecodeMessage unmarshals the protobuf body identified by code.
func DecodeMessage(data []byte, code int16) (*dynamicpb.Message, error) {
	desc, err := getMsgDescByMsgCode(code)
	if err != nil {
		return nil, err
	}

	msg := dynamicpb.NewMessage(desc)
	if err := proto.Unmarshal(data, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// EncodeMessage writes the wire frame: 2-byte length (always 0), 2-byte code, body.
func EncodeMessage(obj protoreflect.ProtoMessage) ([]byte, error) {
	if obj == nil {
		return nil, ErrEmptyData
	}

	code, err := getMsgCodeByMsgName(string(obj.ProtoReflect().Descriptor().Name()))
	if err != nil {
		return nil, err
	}

	body, err := proto.Marshal(obj)
	if err != nil {
		return nil, err
	}

	// First two bytes are a leftover length prefix from the original
	// TCP framing. WebSocket already frames the payload, so they stay 0
	// to remain wire-compatible with the demo client.
	return slices.Concat(
		binary.BigEndian.AppendUint16(nil, 0),
		binary.BigEndian.AppendUint16(nil, uint16(code)),
		body,
	), nil
}

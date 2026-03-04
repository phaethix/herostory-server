package codec

import (
	"encoding/binary"
	"slices"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// DecodeMessage decode byte array by message code
func DecodeMessage(data []byte, code int16) (*dynamicpb.Message, error) {
	if len(data) == 0 {
		return nil, ErrEmptyData
	}

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

// EncodeMessage encode message objects to byte array
func EncodeMessage(obj protoreflect.ProtoMessage) ([]byte, error) {
	if obj == nil {
		return nil, ErrEmptyData
	}

	// encode the length of the message
	msgSizeByteArray := make([]byte, 2)
	binary.BigEndian.PutUint16(msgSizeByteArray, 0)

	// encode the code of the message
	code, err := getMsgCodeByMsgName(string(obj.ProtoReflect().Descriptor().Name()))
	if err != nil {
		return nil, err
	}
	msgCodeByteArray := make([]byte, 2)
	binary.BigEndian.PutUint16(msgCodeByteArray, uint16(code))

	// encode the body of the message
	msgBodyByteArray, err := proto.Marshal(obj)
	if err != nil {
		return nil, err
	}

	return slices.Concat(msgSizeByteArray, msgCodeByteArray, msgBodyByteArray), nil
}

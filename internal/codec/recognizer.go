package codec

import (
	"strings"

	"herostory-server/internal/pb"

	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	msgCodeAndMsgDescMap = make(map[int16]protoreflect.MessageDescriptor)
	msgNameAndMsgCodeMap = make(map[string]int16)
)

// canonicalMsgName maps both enum names (USER_LOGIN_CMD) and message
// names (UserLoginCmd) onto one lookup key.
func canonicalMsgName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, "_", ""))
}

func getMsgDescByMsgCode(code int16) (protoreflect.MessageDescriptor, error) {
	if code < 0 {
		return nil, ErrInvalidMsgCode
	}
	desc, ok := msgCodeAndMsgDescMap[code]
	if !ok {
		return nil, ErrUnknownMessage
	}
	return desc, nil
}

func getMsgCodeByMsgName(name string) (int16, error) {
	if name == "" {
		return -1, ErrEmptyMsgName
	}
	code, ok := msgNameAndMsgCodeMap[canonicalMsgName(name)]
	if !ok {
		return -1, ErrUnknownMessage
	}
	return code, nil
}

// InitMaps builds the code↔descriptor tables. Call once at process start.
func InitMaps() {
	for k, v := range pb.MsgCode_value {
		msgNameAndMsgCodeMap[canonicalMsgName(k)] = int16(v)
	}

	msgDescLst := pb.File_api_proto_game_msg_proto.Messages()
	for i := range msgDescLst.Len() {
		msgDesc := msgDescLst.Get(i)
		if code, ok := msgNameAndMsgCodeMap[canonicalMsgName(string(msgDesc.Name()))]; ok {
			msgCodeAndMsgDescMap[code] = msgDesc
		}
	}
}

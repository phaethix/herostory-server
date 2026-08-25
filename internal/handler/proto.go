package handler

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"
)

// unmarshalCmd projects a wire-decoded dynamic message onto T.
// The codec cannot produce a typed *pb.T because the frame only carries
// a numeric code.
func unmarshalCmd[T any, PT interface {
	*T
	proto.Message
}](src *dynamicpb.Message) *T {
	dst := new(T)
	if src != nil {
		proto.Merge(PT(dst), src)
	}
	return dst
}

package handler

import (
	"testing"

	"herostory-server/internal/pb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"
)

func TestUnmarshalCmd_CopiesFields(t *testing.T) {
	want := &pb.UserMoveToCmd{
		MoveFromPosX: 1,
		MoveFromPosY: 2,
		MoveToPosX:   3,
		MoveToPosY:   4,
	}
	src := dynamicpb.NewMessage(want.ProtoReflect().Descriptor())
	proto.Merge(src, want)

	got := unmarshalCmd[pb.UserMoveToCmd](src)
	require.NotNil(t, got)
	assert.True(t, proto.Equal(want, got))
}

func TestUnmarshalCmd_NilSrc(t *testing.T) {
	got := unmarshalCmd[pb.UserLoginCmd](nil)
	require.NotNil(t, got)
	assert.True(t, proto.Equal(&pb.UserLoginCmd{}, got))
}

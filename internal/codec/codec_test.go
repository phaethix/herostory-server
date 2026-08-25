package codec

import (
	"os"
	"testing"

	"herostory-server/internal/pb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestMain(m *testing.M) {
	InitMaps()
	os.Exit(m.Run())
}

func TestEncodeDecode_RoundTrip(t *testing.T) {
	want := &pb.UserMoveToCmd{
		MoveFromPosX: 10,
		MoveFromPosY: 20,
		MoveToPosX:   30,
		MoveToPosY:   40,
	}
	raw, err := EncodeMessage(want)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(raw), 4)

	got, err := DecodeMessage(raw[4:], int16(pb.MsgCode_USER_MOVE_TO_CMD))
	require.NoError(t, err)

	cmd := new(pb.UserMoveToCmd)
	proto.Merge(cmd, got)
	assert.True(t, proto.Equal(want, cmd))
}

func TestEncodeMessage_Nil(t *testing.T) {
	_, err := EncodeMessage(nil)
	assert.ErrorIs(t, err, ErrEmptyData)
}

func TestDecodeMessage_UnknownCode(t *testing.T) {
	_, err := DecodeMessage(nil, 99)
	assert.ErrorIs(t, err, ErrUnknownMessage)
}

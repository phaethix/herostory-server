package websocket

import (
	"testing"

	"herostory-server/internal/pb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteMsg_QueuesBeforeSendLoopStarts(t *testing.T) {
	ctx := NewCmdContext(nil, 1)

	ctx.WriteMsg(&pb.UserQuitResult{QuitUserId: 7})

	select {
	case got := <-ctx.msgQ:
		require.NotNil(t, got)
		quit, ok := got.(*pb.UserQuitResult)
		require.True(t, ok)
		assert.Equal(t, uint32(7), quit.QuitUserId)
	default:
		t.Fatal("WriteMsg dropped the message because the send queue was not ready")
	}
}

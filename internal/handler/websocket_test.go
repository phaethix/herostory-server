package handler

import (
	"herostory-server/internal/pb"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func Test_WebSocketConnection(t *testing.T) {
	dialer := websocket.Dialer{
		HandshakeTimeout: 3 * time.Second,
	}

	conn, resp, err := dialer.Dial("ws://localhost:12345/websocket", nil)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	assert.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte("Hello World")))
}

func Test_ParseUserLoginCmd(t *testing.T) {
	expectedCmd := &pb.UserLoginCmd{
		UserName: "test_player",
		Password: "test_pass123",
	}

	protoData, err := proto.Marshal(expectedCmd)
	require.NoError(t, err)

	fullData := append([]byte{0x00, 0x01, 0x02, 0x03}, protoData...)

	cmd := &pb.UserLoginCmd{}
	err = proto.Unmarshal(fullData[4:], cmd)

	assert.NoError(t, err)
	assert.Equal(t, expectedCmd.UserName, cmd.UserName)
	assert.Equal(t, expectedCmd.Password, cmd.Password)
}

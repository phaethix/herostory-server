package server

import (
	"cmp"
	"context"
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"herostory-server/internal/codec"
	"herostory-server/internal/database"
	"herostory-server/internal/model"
	"herostory-server/internal/pb"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	codec.InitMaps()
	os.Exit(m.Run())
}

func dial(t *testing.T) *websocket.Conn {
	t.Helper()
	srv := httptest.NewTestServer(t, http.HandlerFunc(WebSocketHandshake))
	tr, ok := srv.Client().Transport.(*http.Transport)
	require.True(t, ok)

	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	// Default net.Dial cannot see NewTestServer's in-memory network.
	conn, resp, err := (&websocket.Dialer{
		NetDialContext: tr.DialContext,
	}).DialContext(ctx, "ws://example.com/", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func requireDB(t *testing.T) {
	t.Helper()
	if database.GetDB() != nil {
		return
	}
	dsn := cmp.Or(os.Getenv("MYSQL_DSN"),
		"root:happycoding@tcp(127.0.0.1:3306)/hero_story?charset=utf8mb4&parseTime=True&loc=Local")
	if err := database.Open(dsn, &gorm.Config{}); err != nil {
		t.Skipf("mysql not available: %v", err)
	}
	if err := database.GetDB().AutoMigrate(&model.User{}); err != nil {
		t.Skipf("auto migrate failed: %v", err)
	}
}

func send(t *testing.T, conn *websocket.Conn, msg proto.Message) {
	t.Helper()
	data, err := codec.EncodeMessage(msg)
	require.NoError(t, err)
	require.NoError(t, conn.WriteMessage(websocket.BinaryMessage, data))
}

func recv(t *testing.T, conn *websocket.Conn) (uint16, []byte) {
	t.Helper()
	deadline, _ := t.Context().Deadline()
	if deadline.IsZero() {
		deadline = time.Now().Add(10 * time.Second)
	}
	_ = conn.SetReadDeadline(deadline)
	_, raw, err := conn.ReadMessage()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(raw), 4)
	return binary.BigEndian.Uint16(raw[2:4]), raw[4:]
}

func recvAs[T any, PT interface {
	*T
	proto.Message
}](t *testing.T, conn *websocket.Conn, wantCode pb.MsgCode) *T {
	t.Helper()
	code, body := recv(t, conn)
	assert.Equal(t, uint16(wantCode), code)
	msg := PT(new(T))
	require.NoError(t, proto.Unmarshal(body, msg))
	return (*T)(msg)
}

func TestWebSocketConnection(t *testing.T) {
	conn := dial(t)
	assert.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte("ping")))
}

func TestParseUserLoginCmd(t *testing.T) {
	want := &pb.UserLoginCmd{UserName: "test_player", Password: "test_pass123"}
	data, err := proto.Marshal(want)
	require.NoError(t, err)

	got := new(pb.UserLoginCmd)
	require.NoError(t, proto.Unmarshal(data, got))
	assert.True(t, proto.Equal(want, got))
}

func TestUserLogin(t *testing.T) {
	requireDB(t)

	for _, tt := range []struct {
		name, user, pass string
	}{
		{"existing_user", "EEE", "123456"},
		{"new_user", "Alice", "abc123"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			conn := dial(t)
			send(t, conn, &pb.UserLoginCmd{UserName: tt.user, Password: tt.pass})

			result := recvAs[pb.UserLoginResult](t, conn, pb.MsgCode_USER_LOGIN_RESULT)
			t.Logf("userId=%d userName=%s heroAvatar=%s", result.UserId, result.UserName, result.HeroAvatar)

			assert.NotZero(t, result.UserId)
			assert.Equal(t, tt.user, result.UserName)
		})
	}
}

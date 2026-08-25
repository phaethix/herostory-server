package handler

import (
	"testing"

	"herostory-server/internal/pb"

	"github.com/stretchr/testify/assert"
)

func TestUserStopHandlerRegistered(t *testing.T) {
	h := CreateCmdHandler(uint16(pb.MsgCode_USER_STOP_CMD))
	assert.NotNil(t, h, "USER_STOP_CMD must be wired to a handler")
}

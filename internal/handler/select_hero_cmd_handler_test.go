package handler

import (
	"testing"

	"herostory-server/internal/pb"

	"github.com/stretchr/testify/assert"
)

func TestSelectHeroHandlerRegistered(t *testing.T) {
	h := CreateCmdHandler(uint16(pb.MsgCode_SELECT_HERO_CMD))
	assert.NotNil(t, h, "SELECT_HERO_CMD must be wired to a handler")
}

package handler

import (
	"testing"

	"herostory-server/internal/pb"
	"herostory-server/internal/rank"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRankHandlerRegistered(t *testing.T) {
	h := CreateCmdHandler(uint16(pb.MsgCode_GET_RANK_CMD))
	assert.NotNil(t, h, "GET_RANK_CMD must be wired to a handler")
}

func TestRankResult_MapsItems(t *testing.T) {
	got := rankResult([]rank.Item{
		{RankID: 1, UserID: 9, UserName: "alice", HeroAvatar: "Hero_A", Win: 3},
	})
	require.Len(t, got.RankItem, 1)
	assert.Equal(t, uint32(1), got.RankItem[0].RankId)
	assert.Equal(t, uint32(9), got.RankItem[0].UserId)
	assert.Equal(t, "alice", got.RankItem[0].UserName)
	assert.Equal(t, "Hero_A", got.RankItem[0].HeroAvatar)
	assert.Equal(t, uint32(3), got.RankItem[0].Win)
}

func TestRankResult_NilIsEmpty(t *testing.T) {
	assert.Empty(t, rankResult(nil).RankItem)
}

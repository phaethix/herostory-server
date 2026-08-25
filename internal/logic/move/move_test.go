package move

import (
	"testing"
	"time"

	"herostory-server/internal/game"
	"herostory-server/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStop_StandsAtDestinationWhenTravelComplete(t *testing.T) {
	u := &game.OnlineUser{
		UserID:     1,
		UserName:   "alice",
		HeroAvatar: model.DefaultHeroAvatar,
		CurrHp:     model.DefaultMaxHp,
		MoveState: &game.MoveState{
			FromPosX:  0,
			FromPosY:  0,
			ToPosX:    300,
			ToPosY:    0,
			StartTime: uint64(time.Now().Add(-2 * time.Second).UnixMilli()),
		},
	}
	game.AddOnlineUser(u)
	t.Cleanup(func() { game.RemoveOnlineUser(u.UserID) })

	result := Stop(u.UserID)

	require.NotNil(t, result)
	assert.Equal(t, uint32(u.UserID), result.StopUserId)
	assert.InDelta(t, 300, result.StopAtPosX, 0.01)
	assert.InDelta(t, 0, result.StopAtPosY, 0.01)

	stopped := game.GetOnlineUser(u.UserID).MoveState
	require.NotNil(t, stopped)
	assert.InDelta(t, 300, stopped.FromPosX, 0.01)
	assert.InDelta(t, 300, stopped.ToPosX, 0.01)
}

func TestStop_IgnoresOfflineUser(t *testing.T) {
	assert.Nil(t, Stop(999))
}

func TestStop_StandsMidwayThroughTravel(t *testing.T) {
	u := &game.OnlineUser{
		UserID:     2,
		UserName:   "bob",
		HeroAvatar: model.DefaultHeroAvatar,
		CurrHp:     model.DefaultMaxHp,
		MoveState: &game.MoveState{
			FromPosX:  0,
			FromPosY:  0,
			ToPosX:    300,
			ToPosY:    0,
			StartTime: uint64(time.Now().Add(-500 * time.Millisecond).UnixMilli()),
		},
	}
	game.AddOnlineUser(u)
	t.Cleanup(func() { game.RemoveOnlineUser(u.UserID) })

	result := Stop(u.UserID)

	require.NotNil(t, result)
	assert.InDelta(t, 150, result.StopAtPosX, 30)
	assert.InDelta(t, 0, result.StopAtPosY, 0.01)
}

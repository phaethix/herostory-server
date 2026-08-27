package attack

import (
	"testing"

	"herostory-server/internal/game"
	"herostory-server/internal/model"
	"herostory-server/internal/pb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func addUser(t *testing.T, id int, hp int32) {
	t.Helper()
	game.AddOnlineUser(&game.OnlineUser{
		UserID:     id,
		UserName:   "u",
		HeroAvatar: model.DefaultHeroAvatar,
		CurrHp:     hp,
	})
	t.Cleanup(func() { game.RemoveOnlineUser(id) })
}

func TestApply_SetsKillWhenTargetDies(t *testing.T) {
	addUser(t, 1, model.DefaultMaxHp)
	addUser(t, 2, subtractHp)

	r := Apply(1, &pb.UserAttkCmd{TargetUserId: 2})
	require.NotNil(t, r)
	assert.Equal(t, 1, r.WinnerID)
	assert.Equal(t, 2, r.LoserID)
}

func TestApply_NoKillWhenTargetLives(t *testing.T) {
	addUser(t, 3, model.DefaultMaxHp)
	addUser(t, 4, model.DefaultMaxHp)

	r := Apply(3, &pb.UserAttkCmd{TargetUserId: 4})
	require.NotNil(t, r)
	assert.Zero(t, r.LoserID)
}

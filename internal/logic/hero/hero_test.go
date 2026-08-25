package hero

import (
	"testing"

	"herostory-server/internal/game"
	"herostory-server/internal/model"
	"herostory-server/internal/pb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApply_UpdatesOnlineUserAvatar(t *testing.T) {
	u := &game.OnlineUser{
		UserID:     1,
		UserName:   "alice",
		HeroAvatar: model.DefaultHeroAvatar,
		CurrHp:     model.DefaultMaxHp,
	}
	game.AddOnlineUser(u)
	t.Cleanup(func() { game.RemoveOnlineUser(u.UserID) })

	result := Apply(u.UserID, &pb.SelectHeroCmd{HeroAvatar: "Hero_Hammer"})

	require.NotNil(t, result)
	assert.Equal(t, "Hero_Hammer", result.HeroAvatar)
	assert.Equal(t, "Hero_Hammer", game.GetOnlineUser(u.UserID).HeroAvatar)
}

func TestApply_IgnoresOfflineUser(t *testing.T) {
	assert.Nil(t, Apply(999, &pb.SelectHeroCmd{HeroAvatar: "Hero_Hammer"}))
}

func TestApply_EmptyAvatarIsFailure(t *testing.T) {
	u := &game.OnlineUser{
		UserID:     2,
		UserName:   "bob",
		HeroAvatar: model.DefaultHeroAvatar,
		CurrHp:     model.DefaultMaxHp,
	}
	game.AddOnlineUser(u)
	t.Cleanup(func() { game.RemoveOnlineUser(u.UserID) })

	result := Apply(u.UserID, &pb.SelectHeroCmd{HeroAvatar: ""})

	require.NotNil(t, result)
	assert.Empty(t, result.HeroAvatar)
	assert.Equal(t, model.DefaultHeroAvatar, game.GetOnlineUser(u.UserID).HeroAvatar)
}

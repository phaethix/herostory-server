package login

import (
	"testing"
	"time"

	"herostory-server/internal/database"
	"herostory-server/internal/model"
	"herostory-server/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func withDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}))
	prev := database.GetDB()
	database.SetDB(db)
	t.Cleanup(func() { database.SetDB(prev) })
}

func insertWounded(t *testing.T, hp int32) *model.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	require.NoError(t, err)
	u := &model.User{
		UserName:   "wounded",
		Password:   string(hash),
		HeroAvatar: model.DefaultHeroAvatar,
		CurrHp:     hp,
		CreateTime: time.Now().Unix(),
	}
	require.NoError(t, repository.CreateUser(u))
	return u
}

func TestSessionHP_DoesNotWriteCurrHp(t *testing.T) {
	withDB(t)
	u := insertWounded(t, 10)

	assert.Equal(t, model.DefaultMaxHp, SessionHP(u))

	got, err := repository.GetUserByName(u.UserName)
	require.NoError(t, err)
	assert.Equal(t, int32(10), got.CurrHp)
}

func TestDoLogin_RepairsWoundedHP(t *testing.T) {
	withDB(t)
	u := insertWounded(t, 10)

	got := doLogin(u.UserName, "secret")
	require.NotNil(t, got)
	assert.Equal(t, model.DefaultMaxHp, got.CurrHp)

	stored, err := repository.GetUserByName(u.UserName)
	require.NoError(t, err)
	assert.Equal(t, model.DefaultMaxHp, stored.CurrHp)
}

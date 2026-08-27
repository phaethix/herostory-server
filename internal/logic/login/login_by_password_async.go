package login

import (
	"context"
	"errors"
	"time"

	"herostory-server/internal/model"
	"herostory-server/internal/rank"
	"herostory-server/internal/repository"
	asyncop "herostory-server/pkg/async_op"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// LoginByPasswordAsync authenticates, or auto-registers if the name is new.
// It returns immediately so bcrypt/DB cannot stall the game loop.
// Empty credentials yield nil.
func LoginByPasswordAsync(username, password string) *asyncop.AsyncBizResult[model.User] {
	if username == "" || password == "" {
		return nil
	}

	bizResult := &asyncop.AsyncBizResult[model.User]{}
	asyncop.Process(
		asyncop.StrToBindID(username),
		func() { bizResult.SetReturnedObj(doLogin(username, password)) },
		nil,
	)
	return bizResult
}

func doLogin(username, password string) *model.User {
	user, err := repository.GetUserByName(username)
	if errors.Is(err, repository.ErrNotFound) {
		u := registerNewUser(username, password)
		writeRankProfile(u)
		return u
	}
	if err != nil {
		log.Error().
			Err(err).
			Str("username", username).
			Msg("query user failed")
		return nil
	}
	if !verifyPassword(user, password) {
		return nil
	}

	updateLastLogin(user)
	log.Info().
		Str("username", username).
		Int("userId", user.ID).
		Msg("user logged in")
	writeRankProfile(user)
	return user
}

func writeRankProfile(u *model.User) {
	if u == nil {
		return
	}
	if err := rank.PutBasicInfo(context.Background(), u.ID, u.UserName, u.HeroAvatar); err != nil {
		log.Error().
			Err(err).
			Int("userId", u.ID).
			Msg("rank: put BasicInfo failed")
	}
}

func registerNewUser(username, password string) *model.User {
	hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if hashErr != nil {
		log.Error().
			Err(hashErr).
			Str("username", username).
			Msg("bcrypt hash failed during registration")
		return nil
	}

	newUser := &model.User{
		UserName:   username,
		Password:   string(hashedPassword),
		HeroAvatar: model.DefaultHeroAvatar,
		CurrHp:     model.DefaultMaxHp,
		CreateTime: time.Now().Unix(),
	}
	if err := repository.CreateUser(newUser); err != nil {
		log.Error().
			Err(err).
			Str("username", username).
			Msg("create user failed")
		return nil
	}

	log.Info().
		Str("username", username).
		Int("userId", newUser.ID).
		Msg("new user registered")
	return newUser
}

func verifyPassword(user *model.User, password string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		log.Warn().
			Str("username", user.UserName).
			Msg("login failed: wrong password")
		return false
	}
	return true
}

func updateLastLogin(user *model.User) {
	if err := repository.UpdateLastLogin(user.ID); err != nil {
		log.Warn().
			Err(err).
			Int("userId", user.ID).
			Msg("update last login time failed")
	}
}

// SessionHP is always full HP. The login/entry/who-else protos carry no
// HP field, so the client always draws a full bar; keeping a wounded
// DB value would look healthy on screen and then die in one hit.
func SessionHP(user *model.User) int32 {
	hp := model.DefaultMaxHp
	if user == nil || user.CurrHp == hp {
		return hp
	}
	if err := repository.UpdateCurrHp(user.ID, hp); err != nil {
		log.Error().
			Err(err).
			Int("userId", user.ID).
			Int32("currHp", hp).
			Msg("login: respawn HP repair failed")
	}
	return hp
}

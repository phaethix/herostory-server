package login

import (
	"errors"
	"herostory-server/internal/model"
	"herostory-server/internal/repository"
	asyncop "herostory-server/pkg/async_op"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// LoginByPasswordAsync performs user authentication (or auto-registration)
// asynchronously. It returns an *asyncop.AsyncBizResult[model.User] immediately
// without blocking the main thread. The caller should register an OnComplete
// callback on the returned result; once the database I/O finishes, the callback
// will be dispatched to the main thread with the result available via GetReturnedObj.
func LoginByPasswordAsync(username, password string) *asyncop.AsyncBizResult[model.User] {
	if username == "" || password == "" {
		return nil
	}

	bizResult := &asyncop.AsyncBizResult[model.User]{}

	asyncop.Process(
		asyncop.StrToBindID(username),
		func() {
			// This closure runs on an async worker goroutine.
			user := doLogin(username, password)
			bizResult.SetReturnedObj(user)
		},
		nil,
	)

	return bizResult
}

// doLogin is the synchronous implementation that runs inside an async worker.
func doLogin(username, password string) *model.User {
	user, err := repository.GetUserByName(username)

	if errors.Is(err, repository.ErrNotFound) {
		return registerNewUser(username, password)
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

	return user
}

// registerNewUser creates a new user account with the supplied password.
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

// verifyPassword checks whether the supplied password matches the stored hash.
func verifyPassword(user *model.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Warn().
			Str("username", user.UserName).
			Msg("login failed: wrong password")
		return false
	}
	return true
}

// updateLastLogin updates the user's last login timestamp (non-critical).
func updateLastLogin(user *model.User) {
	if err := repository.UpdateLastLogin(user.ID); err != nil {
		log.Warn().
			Err(err).
			Int("userId", user.ID).
			Msg("update last login time failed")
	}
}

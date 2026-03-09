package login

import (
	"herostory-server/internal/model"
	"herostory-server/internal/repository"
	"herostory-server/pkg/main_thread"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// LoginByPasswordAsync performs user authentication (or auto-registration) in a
// background goroutine so that the game main thread is never blocked by database
// I/O. when the operation completes, callback is invoked on the main thread with
// the authenticated *model.User, or nil on failure.
func LoginByPasswordAsync(username, password string, callback func(user *model.User)) {
	if username == "" || password == "" {
		main_thread.Process(func() { callback(nil) })
		return
	}

	go func() {
		user := doLogin(username, password)
		// deliver result back to the main thread
		main_thread.Process(func() { callback(user) })
	}()
}

// doLogin is the synchronous implementation that runs inside a goroutine.
func doLogin(username, password string) *model.User {
	// try to get the user by name
	user, err := repository.GetUserByName(username)

	if err == repository.ErrNotFound {
		// user does not exist – auto-register with the supplied password
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

		if err = repository.CreateUser(newUser); err != nil {
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

	if err != nil {
		log.Error().
			Err(err).
			Str("username", username).
			Msg("query user failed")
		return nil
	}

	// user exists – verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Warn().
			Str("username", username).
			Msg("login failed: wrong password")
		return nil
	}

	// update last login timestamp (non-critical, log but don't fail)
	err = repository.UpdateLastLogin(user.ID)
	if err != nil {
		log.Warn().
			Err(err).
			Int("userId", user.ID).
			Msg("update last login time failed")
	}

	log.Info().
		Str("username", username).
		Int("userId", user.ID).
		Msg("user logged in")

	return user
}

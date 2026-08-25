package repository

import (
	"errors"
	"time"

	"herostory-server/internal/database"
	"herostory-server/internal/model"

	"gorm.io/gorm"
)

var (
	ErrNotFound               = errors.New("record not found")
	ErrDatabaseNotInitialized = errors.New("database not initialized")
)

func db() (*gorm.DB, error) {
	d := database.GetDB()
	if d == nil {
		return nil, ErrDatabaseNotInitialized
	}
	return d, nil
}

// GetUserByName returns ErrNotFound when the username does not exist.
func GetUserByName(username string) (*model.User, error) {
	d, err := db()
	if err != nil {
		return nil, err
	}

	var user model.User
	res := d.Where("user_name = ?", username).First(&user)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, res.Error
	}
	return &user, nil
}

// CreateUser expects Password to already be a bcrypt hash.
func CreateUser(u *model.User) error {
	d, err := db()
	if err != nil {
		return err
	}
	return d.Create(u).Error
}

func UpdateLastLogin(userID int) error {
	d, err := db()
	if err != nil {
		return err
	}
	return d.Model(&model.User{}).
		Where("id = ?", userID).
		Update("last_login_time", time.Now().Unix()).
		Error
}

func UpdateCurrHp(userID int, currHp int32) error {
	d, err := db()
	if err != nil {
		return err
	}
	return d.Model(&model.User{}).
		Where("id = ?", userID).
		Update("curr_hp", currHp).
		Error
}

func UpdateHeroAvatar(userID int, heroAvatar string) error {
	d, err := db()
	if err != nil {
		return err
	}
	return d.Model(&model.User{}).
		Where("id = ?", userID).
		Update("hero_avatar", heroAvatar).
		Error
}

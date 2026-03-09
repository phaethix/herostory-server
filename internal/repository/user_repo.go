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

// GetUserByName retrieves a user by username.
// returns ErrNotFound if the user does not exist.
func GetUserByName(username string) (*model.User, error) {
	db := database.GetDB()
	if db == nil {
		return nil, ErrDatabaseNotInitialized
	}

	var user model.User
	res := db.Where("user_name = ?", username).First(&user)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, res.Error
	}

	return &user, nil
}

// CreateUser inserts a new user record.
// the caller should ensure the password field is already hashed and that UserName is unique.
func CreateUser(u *model.User) error {
	db := database.GetDB()
	if db == nil {
		return ErrDatabaseNotInitialized
	}

	return db.Create(u).Error
}

// UpdateLastLogin updates the last_login_time for the given user id.
func UpdateLastLogin(userID int) error {
	db := database.GetDB()
	if db == nil {
		return ErrDatabaseNotInitialized
	}

	now := time.Now().Unix()
	return db.Model(&model.User{}).
		Where("id = ?", userID).
		Update("last_login_time", now).
		Error
}

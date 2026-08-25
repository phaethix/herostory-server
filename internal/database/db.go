package database

import (
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	maxOpenConns    = 128
	maxIdleConns    = 16
	connMaxLifetime = 2 * time.Minute
)

var (
	mu sync.RWMutex
	DB *gorm.DB
)

func SetDB(db *gorm.DB) {
	mu.Lock()
	defer mu.Unlock()
	DB = db
}

// GetDB returns the connection set by Open, or nil if Open has not succeeded.
func GetDB() *gorm.DB {
	mu.RLock()
	defer mu.RUnlock()
	return DB
}

// Open connects to MySQL, pings it, and stores the handle for GetDB.
func Open(dsn string, config *gorm.Config) error {
	db, err := gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	if err := sqlDB.Ping(); err != nil {
		return err
	}

	SetDB(db)
	return nil
}

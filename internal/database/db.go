package database

import (
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	// mu protects access to the global DB variable.
	mu sync.RWMutex

	// DB is the shared *gorm.DB instance used by repository packages. It must
	// be initialized during application startup (typically in bootstrap.InitApp)
	// before any repository call.
	DB *gorm.DB
)

// SetDB stores the provided *gorm.DB in a threadsafe manner. It is usually
// called once during start-up and never changed after that.
func SetDB(db *gorm.DB) {
	mu.Lock()
	defer mu.Unlock()
	DB = db
}

// GetDB returns the database connection previously set via SetDB, or nil if
// the connection has not been initialized yet.
func GetDB() *gorm.DB {
	mu.RLock()
	defer mu.RUnlock()
	return DB
}

// Open establishes a connection to a MySQL database using the given DSN and
// optional GORM configuration. On success it stores the resulting *gorm.DB in
// the global variable so that callers may subsequently call GetDB().
// It also configures connection pool settings and tests the connection.
func Open(dsn string, config *gorm.Config) error {
	db, err := gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		return err
	}

	// configure connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// maximum number of open connections
	sqlDB.SetMaxOpenConns(128)
	// maximum number of idle connections
	sqlDB.SetMaxIdleConns(16)
	// maximum connection lifetime
	sqlDB.SetConnMaxLifetime(2 * time.Minute)

	// test the connection
	if err := sqlDB.Ping(); err != nil {
		return err
	}

	SetDB(db)
	return nil
}

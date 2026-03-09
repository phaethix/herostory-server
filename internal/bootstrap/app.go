package bootstrap

import (
	"os"

	"herostory-server/internal/codec"
	"herostory-server/internal/database"
	"herostory-server/internal/model"
	"herostory-server/pkg/logger"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func InitApp() {
	logger.InitZeroLogger("./storage/logs", "biz_server")
	codec.InitMaps()

	// initialize database connection
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		// fallback to a default (developer convenience)
		dsn = "root:happycoding@tcp(127.0.0.1:3306)/hero_story?charset=utf8mb4&parseTime=True&loc=Local"
	}

	if err := database.Open(dsn, &gorm.Config{}); err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to open database")
	}

	// auto migrate schema
	db := database.GetDB()
	if err := db.AutoMigrate(&model.User{}); err != nil {
		log.Fatal().
			Err(err).
			Msg("auto migrate failed")
	}
}

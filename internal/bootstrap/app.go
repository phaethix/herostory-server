package bootstrap

import (
	"cmp"
	"os"

	"herostory-server/internal/codec"
	"herostory-server/internal/database"
	"herostory-server/internal/model"
	"herostory-server/internal/rank"
	"herostory-server/pkg/logger"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

const (
	defaultMySQLDSN  = "root:happycoding@tcp(127.0.0.1:3306)/hero_story?charset=utf8mb4&parseTime=True&loc=Local"
	defaultRedisAddr = "127.0.0.1:6379"
)

// InitApp sets up logging, codec maps, MySQL, and Redis. It fatals on store failure.
func InitApp() {
	logger.InitZeroLogger("./storage/logs", "biz_server")
	codec.InitMaps()

	dsn := cmp.Or(os.Getenv("MYSQL_DSN"), defaultMySQLDSN)
	if err := database.Open(dsn, &gorm.Config{}); err != nil {
		log.Fatal().Err(err).Msg("failed to open database")
	}

	if err := database.GetDB().AutoMigrate(&model.User{}); err != nil {
		log.Fatal().Err(err).Msg("auto migrate failed")
	}

	if err := rank.Open(cmp.Or(os.Getenv("REDIS_ADDR"), defaultRedisAddr)); err != nil {
		log.Fatal().Err(err).Msg("failed to open redis")
	}
}

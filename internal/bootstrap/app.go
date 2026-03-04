package bootstrap

import (
	"herostory-server/internal/codec"
	"herostory-server/pkg/logger"
)

func InitApp() {
	logger.InitZeroLogger("./storage/logs", "biz_server")
	codec.InitMaps()
}

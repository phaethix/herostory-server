package logger

import (
	"fmt"
	"io"
	stdlog "log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

var formatTime = func(i any) string {
	return fmt.Sprintf("%+v", i)
}

var formatCaller = func(i any) string {
	caller, ok := i.(string)
	if !ok {
		return ""
	}
	_, file, found := strings.CutLast(caller, "/")
	if found {
		return file + ":"
	}
	return caller + ":"
}

// InitZeroLogger writes to stdout and path/filename.log.
func InitZeroLogger(path, filename string) {
	zerolog.TimeFieldFormat = "2006/01/02 15:04:05.000000"
	zerolog.FormattedLevels = map[zerolog.Level]string{
		zerolog.TraceLevel: "TRACE",
		zerolog.DebugLevel: "DEBUG",
		zerolog.InfoLevel:  "INFO",
		zerolog.WarnLevel:  "WARN",
		zerolog.ErrorLevel: "ERROR",
		zerolog.FatalLevel: "FATAL",
		zerolog.PanicLevel: "PANIC",
	}

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		stdlog.Fatalf("create log dir failed: %v", err)
	}

	file := &lumberjack.Logger{
		Filename:   filepath.Join(path, filename+".log"),
		MaxSize:    10,
		MaxAge:     0,
		MaxBackups: 30,
		Compress:   true,
	}

	log.Logger = zerolog.New(io.MultiWriter(
		zerolog.ConsoleWriter{
			Out:             os.Stdout,
			FormatTimestamp: formatTime,
			FormatCaller:    formatCaller,
		},
		zerolog.ConsoleWriter{
			Out:             file,
			NoColor:         true,
			FormatTimestamp: formatTime,
			FormatCaller:    formatCaller,
		},
	)).With().Caller().Timestamp().Logger()

	stdlog.SetOutput(io.MultiWriter(os.Stdout, file))
}

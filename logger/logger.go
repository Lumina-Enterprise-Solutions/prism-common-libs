package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init initializes the global zerolog logger for structured JSON logging.
func Init() {
	logLevelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))
	var level zerolog.Level

	switch logLevelStr {
	case "debug":
		level = zerolog.DebugLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	default:
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)

	// Gunakan output konsol yang lebih mudah dibaca saat TTY aktif (pengembangan lokal),
	// dan JSON untuk lingkungan lain (seperti di dalam Docker).
	if isTTY() {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	} else {
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}
}

func isTTY() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

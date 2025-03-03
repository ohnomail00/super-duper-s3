package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func Init(cfgLevel string) {
	var level = slog.LevelDebug
	var addSource = false
	switch cfgLevel {
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		addSource = true
	}
	w := os.Stderr
	slog.SetDefault(slog.New(tint.NewHandler(w, &tint.Options{
		Level:      level,
		AddSource:  addSource,
		TimeFormat: time.DateTime,
	})))

}

package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/dpotapov/slogpfx"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"

	"github.com/MunifTanjim/stremthru/internal/config"
)

var _ = func() *slog.Logger {
	w := os.Stderr

	var handler slog.Handler

	if config.LogFormat == "json" {
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: config.LogLevel,
		})
	} else {
		handler = slogpfx.NewHandler(
			tint.NewHandler(w, &tint.Options{
				Level:      config.LogLevel,
				NoColor:    !isatty.IsTerminal(w.Fd()),
				TimeFormat: time.DateTime,
			}),
			&slogpfx.HandlerOptions{
				PrefixKeys: []string{"scope"},
			},
		)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelInfo)
	return logger
}()

func Scoped(scope string) *slog.Logger {
	return slog.With("scope", scope)
}

package logger

import (
	"log/slog"
	"os"
)

func Init(appEnv string) {
	var handler slog.Handler

	if appEnv == "production" {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}

	slog.SetDefault(slog.New(handler))
}

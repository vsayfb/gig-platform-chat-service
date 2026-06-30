package logger

import (
	"log/slog"
	"os"
)

func Init(appEnv string) {
	var handler slog.Handler

	if appEnv == "local" {
		handler = slog.NewTextHandler(os.Stdout, nil)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	}

	slog.SetDefault(slog.New(handler))
}

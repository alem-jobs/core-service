package main

import (
	"log/slog"
	"os"

	"github.com/aidosgal/alem.core-service/internal/config"
	"github.com/aidosgal/alem.core-service/internal/app"
)

func main() {
	cfg := config.MustLoad()

	log := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		),
	)

	log.Info("starting application")

	application := app.NewServer(cfg, log)

	if err := application.Run(); err != nil {
		panic(err)
	}
}

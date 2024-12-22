package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/SarunasBucius/nutri-price-server/internal/setup"
	"github.com/SarunasBucius/nutri-price-server/migrations"
)

func main() {
	ctx := context.Background()

	config, err := setup.LoadConfig(ctx)
	if err != nil {
		slog.Error("load config", "error", err)
		return
	}

	if err := migrations.MigrateDB(ctx, config.DBPool); err != nil {
		slog.Error("migrate db", "error", err)
		return
	}

	r := setup.LoadRouter(config)

	slog.InfoContext(ctx, "Listening...", "port", config.Port)

	http.ListenAndServe(config.Port, r)
}

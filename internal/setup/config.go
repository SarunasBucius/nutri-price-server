package setup

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type Config struct {
	Port   string
	DBPool *pgxpool.Pool
}

func LoadConfig(ctx context.Context) (Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.WarnContext(ctx, "load .env file", "error", err)
		slog.InfoContext(ctx, "continue in case variables are set without .env file")
	}

	dbPool, err := initPostgres(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return Config{}, fmt.Errorf("init db: %w", err)
	}

	port := os.Getenv("PORT")
	if len(port) == 0 {
		return Config{}, fmt.Errorf("empty port")
	}

	return Config{
		Port:   port,
		DBPool: dbPool,
	}, nil
}

func initPostgres(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}
	return dbPool, nil
}

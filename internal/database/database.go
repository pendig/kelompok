package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Open(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	if databaseURL == "" {
		return nil, errors.New("KELOMPOK_DATABASE_URL is required")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 5
	config.MinConns = 0
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute
	config.HealthCheckPeriod = time.Minute

	return pgxpool.NewWithConfig(ctx, config)
}

func Ping(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return errors.New("database pool is not initialized")
	}
	return pool.Ping(ctx)
}

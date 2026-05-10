package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolSettings struct {
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

func Open(ctx context.Context, databaseURL string, settings PoolSettings) (*pgxpool.Pool, error) {
	if databaseURL == "" {
		return nil, errors.New("KELOMPOK_DATABASE_URL is required")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	config.MaxConns = settings.MaxConns
	config.MinConns = settings.MinConns
	config.MaxConnLifetime = settings.MaxConnLifetime
	config.MaxConnIdleTime = settings.MaxConnIdleTime
	config.HealthCheckPeriod = settings.HealthCheckPeriod

	return pgxpool.NewWithConfig(ctx, config)
}

func Ping(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return errors.New("database pool is not initialized")
	}
	return pool.Ping(ctx)
}

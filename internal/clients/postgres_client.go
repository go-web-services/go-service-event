package clients

import (
	"context"
	"fmt"

	"github.com/go-web-services/go-service-event/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresClient(cfg config.PostgresConfig) (*pgxpool.Pool, error) {
	dsn := cfg.GetDSN()

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	poolConfig.MaxConns = 20
	poolConfig.MinConns = 5

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

package db

import (
	"context"
	"fmt"
	"github.com/binsabit/jetinno-kapsi/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/url"
	"time"
)

type Database struct {
	db *pgxpool.Pool
}

func New(ctx context.Context) (*Database, error) {
	cfg := config.AppConfig
	dsn := url.URL{
		Scheme: cfg.DB_DRIVER,
		User:   url.UserPassword(cfg.DB_USER, cfg.DB_PASSWORD),
		Host:   fmt.Sprintf("%s:%s", cfg.DB_HOST, cfg.DB_PORT),
		Path:   cfg.DB_NAME,
	}

	q := dsn.Query()

	q.Add("sslmode", "disabled")

	dsn.RawQuery = q.Encode()

	poolConfig, err := pgxpool.ParseConfig(dsn.String())
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = 15
	poolConfig.MaxConnIdleTime = time.Minute * 10

	pgxPool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	if err := pgxPool.Ping(ctx); err != nil {
		return nil, err
	}
	return &Database{db: pgxPool}, nil
}

func (d *Database) GetDB() *pgxpool.Pool {
	return d.db
}

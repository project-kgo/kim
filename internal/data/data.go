package data

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

const (
	defaultRedisPingTimeout = 3 * time.Second
	postgresDriverName      = "pgx"
)

type Data struct {
	Redis  *redis.Client
	DB     *sqlx.DB
	logger *slog.Logger
}

func New(redisDSN string, dbDSN string, logger *slog.Logger) (*Data, error) {
	redisDSN = strings.TrimSpace(redisDSN)
	if redisDSN == "" {
		return nil, errors.New("redis dsn is required")
	}
	dbDSN = strings.TrimSpace(dbDSN)
	if dbDSN == "" {
		return nil, errors.New("db dsn is required")
	}

	opts, err := redis.ParseURL(redisDSN)
	if err != nil {
		return nil, fmt.Errorf("parse redis dsn: %w", err)
	}

	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), defaultRedisPingTimeout)
	defer cancel()
	if err = client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	db, err := sqlx.Open(postgresDriverName, dbDSN)
	if err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	if logger != nil {
		logger.Info("data clients initialized",
			slog.String("redis_addr", opts.Addr),
			slog.Int("redis_db", opts.DB),
		)
	}
	return &Data{Redis: client, DB: db, logger: logger}, nil
}

func (d *Data) Close() error {
	if d == nil {
		return nil
	}
	var errs []error
	if d.Redis != nil {
		if err := d.Redis.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close redis: %w", err))
		}
	}
	if d.DB != nil {
		if err := d.DB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close postgres: %w", err))
		}
	}
	if d.logger != nil {
		d.logger.Info("data clients closed")
	}
	return errors.Join(errs...)
}

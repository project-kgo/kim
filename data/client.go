package data

import (
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const postgresDriverName = "pgx"

type Config struct {
	RedisDSN string
	DBDSN    string
}

type Client struct {
	Redis *redis.Client
	DB    *sqlx.DB
}

func NewClient(cfg Config) (*Client, error) {
	redisDSN := strings.TrimSpace(cfg.RedisDSN)
	if redisDSN == "" {
		return nil, errors.New("redis dsn is required")
	}
	dbDSN := strings.TrimSpace(cfg.DBDSN)
	if dbDSN == "" {
		return nil, errors.New("db dsn is required")
	}

	redisOptions, err := redis.ParseURL(redisDSN)
	if err != nil {
		return nil, err
	}

	redisClient := redis.NewClient(redisOptions)
	db, err := sqlx.Open(postgresDriverName, dbDSN)
	if err != nil {
		_ = redisClient.Close()
		return nil, err
	}

	return &Client{
		Redis: redisClient,
		DB:    db,
	}, nil
}

func (c *Client) Close() error {
	if c == nil {
		return nil
	}

	var closeErr error
	if c.Redis != nil {
		closeErr = errors.Join(closeErr, c.Redis.Close())
	}
	if c.DB != nil {
		closeErr = errors.Join(closeErr, c.DB.Close())
	}
	return closeErr
}

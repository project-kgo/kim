package gateway

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	kimgatev1 "github.com/project-kgo/kim-gate/proto/kimgate/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

type ServiceClient interface {
	kimgatev1.GatewayServiceClient
}

type Config struct {
	GatewayService string
	GatewayTimeout time.Duration
}

type Client struct {
	conn    *grpc.ClientConn
	service ServiceClient
}

func NewClient(cfg Config, logger *slog.Logger, resolverBuilder resolver.Builder) (*Client, error) {
	serviceName := strings.TrimSpace(cfg.GatewayService)
	if serviceName == "" {
		return nil, errors.New("gateway service name is required")
	}
	if resolverBuilder == nil {
		return nil, errors.New("resolver builder is required")
	}
	timeout := cfg.GatewayTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	conn, err := grpc.NewClient(
		"etcd:///"+serviceName,
		grpc.WithResolvers(resolverBuilder),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: timeout,
		}),
	)
	if err != nil {
		return nil, err
	}

	if logger != nil {
		logger.Info("gateway client created",
			slog.String("service", serviceName),
			slog.Duration("timeout", timeout),
		)
	}

	return &Client{
		conn:    conn,
		service: kimgatev1.NewGatewayServiceClient(conn),
	}, nil
}

func (c *Client) Service() ServiceClient {
	if c == nil {
		return nil
	}
	return c.service
}

func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) Ready(ctx context.Context) error {
	if c == nil || c.conn == nil {
		return errors.New("gateway client is nil")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	c.conn.Connect()
	return nil
}

package gateway

import (
	"context"
	"errors"
	"strings"

	kimgatev1 "github.com/project-kgo/kim-gate/proto/kimgate/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ServiceClient interface {
	kimgatev1.GatewayServiceClient
}

type Config struct {
	SocketPath string
}

type Client struct {
	conn    *grpc.ClientConn
	service ServiceClient
}

func NewClient(cfg Config) (*Client, error) {
	socketPath := strings.TrimSpace(cfg.SocketPath)
	if socketPath == "" {
		return nil, errors.New("gateway socket path is required")
	}

	conn, err := grpc.NewClient(
		UnixTarget(socketPath),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:    conn,
		service: kimgatev1.NewGatewayServiceClient(conn),
	}, nil
}

func UnixTarget(socketPath string) string {
	return "unix://" + strings.TrimSpace(socketPath)
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
		return errors.New("gateway connection is required")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	c.conn.Connect()
	return nil
}

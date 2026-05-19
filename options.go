package kim

import (
	"errors"
	"strings"

	"github.com/project-kgo/kim/data"
)

const (
	DefaultRoutePrefix   = "/kim"
	DefaultGatewaySocket = "/tmp/kim-gate.sock"
)

type Options struct {
	RoutePrefix   string
	GatewaySocket string
	RedisDSN      string
	DBDSN         string
	DataClient    *data.Client
}

type Option func(*Options) error

func WithRoutePrefix(prefix string) Option {
	return func(opts *Options) error {
		opts.RoutePrefix = strings.TrimSpace(prefix)
		return nil
	}
}

func WithGatewaySocket(path string) Option {
	return func(opts *Options) error {
		opts.GatewaySocket = strings.TrimSpace(path)
		return nil
	}
}

func WithRedisDSN(dsn string) Option {
	return func(opts *Options) error {
		opts.RedisDSN = strings.TrimSpace(dsn)
		return nil
	}
}

func WithDBDSN(dsn string) Option {
	return func(opts *Options) error {
		opts.DBDSN = strings.TrimSpace(dsn)
		return nil
	}
}

func WithDataClient(client *data.Client) Option {
	return func(opts *Options) error {
		if client == nil {
			return errors.New("kim data client is required")
		}
		opts.DataClient = client
		return nil
	}
}

func newOptions(options ...Option) (Options, error) {
	opts := Options{
		RoutePrefix:   DefaultRoutePrefix,
		GatewaySocket: DefaultGatewaySocket,
	}

	for _, option := range options {
		if option == nil {
			return Options{}, errors.New("kim option is required")
		}
		if err := option(&opts); err != nil {
			return Options{}, err
		}
	}

	if err := opts.validate(); err != nil {
		return Options{}, err
	}
	return opts, nil
}

func (o Options) validate() error {
	if o.RoutePrefix == "" {
		return errors.New("kim route prefix is required")
	}
	if !strings.HasPrefix(o.RoutePrefix, "/") {
		return errors.New("kim route prefix must start with /")
	}
	if o.GatewaySocket == "" {
		return errors.New("kim gateway socket is required")
	}
	if o.DataClient != nil {
		return nil
	}
	if o.RedisDSN == "" {
		return errors.New("kim redis dsn is required")
	}
	if o.DBDSN == "" {
		return errors.New("kim db dsn is required")
	}
	return nil
}

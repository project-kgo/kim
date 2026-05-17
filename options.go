package kim

import (
	"errors"
	"strings"
)

const (
	DefaultRoutePrefix   = "/kim"
	DefaultGatewaySocket = "/tmp/kim-gate.sock"
)

type Options struct {
	RoutePrefix   string
	GatewaySocket string
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
	return nil
}

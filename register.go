package kim

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/project-kgo/kim/data"
)

type registrar struct {
	options Options
	data    *data.Client
}

func Register(h *server.Hertz, opts ...Option) error {
	if h == nil {
		return errors.New("hertz server is required")
	}

	options, err := newOptions(opts...)
	if err != nil {
		return err
	}

	dataClient := options.DataClient
	ownsDataClient := dataClient == nil
	if ownsDataClient {
		dataClient, err = data.NewClient(data.Config{
			RedisDSN: options.RedisDSN,
			DBDSN:    options.DBDSN,
		})
		if err != nil {
			return err
		}
	}

	registrar := &registrar{
		options: options,
		data:    dataClient,
	}
	registrar.registerRoutes(h.Group(options.RoutePrefix))
	if ownsDataClient {
		h.Engine.OnShutdown = append(h.Engine.OnShutdown, func(ctx context.Context) {
			_ = ctx
			_ = dataClient.Close()
		})
	}
	return nil
}

func (r *registrar) registerRoutes(group *route.RouterGroup) {
	_ = r
	_ = group
}

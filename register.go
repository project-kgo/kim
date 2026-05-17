package kim

import (
	"errors"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/route"
)

type registrar struct {
	options Options
}

func Register(h *server.Hertz, opts ...Option) error {
	if h == nil {
		return errors.New("hertz server is required")
	}

	options, err := newOptions(opts...)
	if err != nil {
		return err
	}

	registrar := &registrar{
		options: options,
	}
	registrar.registerRoutes(h.Group(options.RoutePrefix))
	return nil
}

func (r *registrar) registerRoutes(group *route.RouterGroup) {
	_ = r
	_ = group
}

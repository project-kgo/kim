package app

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	hertzserver "github.com/cloudwego/hertz/pkg/app/server"
	"github.com/project-kgo/kim/internal/config"
	"github.com/project-kgo/kim/internal/data"
	"github.com/project-kgo/kim/internal/gateway"
	"github.com/project-kgo/kim/internal/handler"
	"github.com/project-kgo/kim/internal/router"
	"github.com/project-kgo/kim/internal/rpc"
)

type App struct {
	cfg        config.Config
	logger     *slog.Logger
	http       *hertzserver.Hertz
	rpcServer  *rpc.Server
	data       *data.Data
	gateway    *gateway.Client
	done       chan error
	once       sync.Once
}

func New(cfg config.Config, logger *slog.Logger, data *data.Data, gatewayClient *gateway.Client, rpcServer *rpc.Server) *App {
	if logger == nil {
		logger = slog.Default()
	}
	http := hertzserver.New(hertzserver.WithHostPorts(cfg.HTTPAddr))
	h := handler.New(logger)
	router.Register(http, h, logger, cfg.RoutePrefix)
	return &App{
		cfg:       cfg,
		logger:    logger,
		http:      http,
		rpcServer: rpcServer,
		data:      data,
		gateway:   gatewayClient,
		done:      make(chan error, 3),
	}
}

func (a *App) Start() error {
	if a == nil {
		return errors.New("app is nil")
	}
	go func() {
		a.logger.Info("hertz server started",
			slog.String("addr", a.cfg.HTTPAddr),
		)
		a.done <- a.http.Run()
	}()
	if a.rpcServer != nil {
		a.rpcServer.Start()
		go func() {
			if err := <-a.rpcServer.Done(); err != nil {
				a.done <- err
			}
		}()
	}
	return nil
}

func (a *App) Done() <-chan error {
	if a == nil {
		return nil
	}
	return a.done
}

func (a *App) Shutdown(ctx context.Context) error {
	if a == nil {
		return nil
	}
	var err error
	a.once.Do(func() {
		httpErr := a.http.Shutdown(ctx)
		var rpcErr error
		if a.rpcServer != nil {
			rpcErr = a.rpcServer.Shutdown(ctx)
		}
		gwErr := a.gateway.Close()
		dataErr := a.data.Close()
		err = errors.Join(httpErr, rpcErr, gwErr, dataErr)
	})
	return err
}

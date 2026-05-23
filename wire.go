//go:build wireinject
// +build wireinject

package main

import (
	"log/slog"

	"github.com/google/wire"
	"github.com/project-kgo/kim/internal/app"
	"github.com/project-kgo/kim/internal/config"
	"github.com/project-kgo/kim/internal/data"
	"github.com/project-kgo/kim/internal/discovery"
	etcddisc "github.com/project-kgo/kim/internal/discovery/etcd"
	etcdreg "github.com/project-kgo/kim/internal/discovery/etcd"
	"github.com/project-kgo/kim/internal/gateway"
	"github.com/project-kgo/kim/internal/rpc"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func Initialize(cfg config.Config, logger *slog.Logger) (*app.App, error) {
	wire.Build(
		ProvideData,
		ProvideEtcdClient,
		ProvideEtcdRegistry,
		etcddisc.ResolverBuilder,
		ProvideGatewayConfig,
		gateway.NewClient,
		ProvideRPCServer,
		app.New,
	)
	return nil, nil
}

func ProvideData(cfg config.Config, logger *slog.Logger) (*data.Data, error) {
	return data.New(cfg.RedisDSN, cfg.DBDSN, logger)
}

func ProvideEtcdClient(cfg config.Config) (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{
		Endpoints: cfg.ETCDEndpoints,
		Username:  cfg.ETCDUsername,
		Password:  cfg.ETCDPassword,
	})
}

func ProvideEtcdRegistry(cli *clientv3.Client, cfg config.Config) discovery.ServiceRegistry {
	return etcdreg.New(cli, cfg.ETCDTTL)
}

func ProvideGatewayConfig(cfg config.Config) gateway.Config {
	return gateway.Config{
		GatewayService: cfg.GatewayServiceName(),
		GatewayTimeout: cfg.GatewayTimeout,
	}
}

func ProvideRPCServer(cfg config.Config, logger *slog.Logger, registry discovery.ServiceRegistry) (*rpc.Server, error) {
	return rpc.NewServer(cfg, nil, logger, registry)
}

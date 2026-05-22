package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	DefaultHTTPAddr        = ":8080"
	DefaultRoutePrefix     = "/kim"
	DefaultGRPCAddr        = ":9090"
	DefaultETCDServiceName = "kim"
	DefaultGatewayService  = "kim-gate"
	DefaultGatewayTimeout  = 5 * time.Second
	DefaultETCDEndpoints   = "localhost:2379"
	DefaultETCDUsername    = ""
	DefaultETCDPassword    = ""
	DefaultETCDTTL         = 15 * time.Second
	DefaultRedisDSN        = "redis://localhost:6379/0"
	DefaultDBDSN           = "postgres://kim:secret@localhost:5432/kim?sslmode=disable"
	DefaultEnv             = ""
	DefaultShutdownTimeout = 10 * time.Second
)

type Config struct {
	HTTPAddr         string
	RoutePrefix      string
	GRPCAddr         string
	ETCDServiceName  string
	GatewayService   string
	GatewayTimeout   time.Duration
	ETCDEndpointsStr string
	ETCDEndpoints    []string
	ETCDUsername     string
	ETCDPassword     string
	ETCDTTL          time.Duration
	RedisDSN         string
	DBDSN            string
	Env              string
	ShutdownTimeout  time.Duration
}

func Load(args []string) (Config, error) {
	v := viper.New()
	setDefaults(v)
	bindEnv(v)

	fs := pflag.NewFlagSet("kim", pflag.ContinueOnError)
	fs.String("config", "", "config file path")
	fs.String("http-addr", "", "http listen address")
	fs.String("route-prefix", "", "http route prefix")
	fs.String("grpc-addr", "", "grpc listen address")
	fs.String("grpc-service", "", "grpc service name for etcd registration")
	fs.String("gateway-service", "", "etcd service name for kim-gate")
	fs.Duration("gateway-timeout", 0, "gateway connection timeout")
	fs.String("etcd-endpoints", "", "comma-separated etcd endpoints")
	fs.String("etcd-username", "", "etcd username")
	fs.String("etcd-password", "", "etcd password")
	fs.Duration("etcd-ttl", 0, "etcd lease ttl")
	fs.String("redis-dsn", "", "redis connection dsn")
	fs.String("db-dsn", "", "postgres connection dsn")
	fs.String("env", "", "deployment environment")
	fs.Duration("shutdown-timeout", 0, "graceful shutdown timeout")
	if err := fs.Parse(normalizeFlagArgs(args)); err != nil {
		return Config{}, err
	}
	if err := bindFlags(v, fs); err != nil {
		return Config{}, err
	}
	if err := readConfigFile(v, fs.Lookup("config").Value.String()); err != nil {
		return Config{}, err
	}

	cfg := Config{
		HTTPAddr:         v.GetString("http.addr"),
		RoutePrefix:      v.GetString("http.route_prefix"),
		GRPCAddr:         v.GetString("grpc.addr"),
		ETCDServiceName:  v.GetString("grpc.service"),
		GatewayService:   v.GetString("grpc.gateway_service"),
		GatewayTimeout:   v.GetDuration("grpc.gateway_timeout"),
		ETCDEndpointsStr: v.GetString("etcd.endpoints"),
		ETCDUsername:     v.GetString("etcd.username"),
		ETCDPassword:     v.GetString("etcd.password"),
		ETCDTTL:          v.GetDuration("etcd.ttl"),
		RedisDSN:         v.GetString("redis.dsn"),
		DBDSN:            v.GetString("db.dsn"),
		Env:              v.GetString("env"),
		ShutdownTimeout:  v.GetDuration("shutdown.timeout"),
	}

	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Defaults() Config {
	return Config{
		HTTPAddr:         DefaultHTTPAddr,
		RoutePrefix:      DefaultRoutePrefix,
		GRPCAddr:         DefaultGRPCAddr,
		ETCDServiceName:  DefaultETCDServiceName,
		GatewayService:   DefaultGatewayService,
		GatewayTimeout:   DefaultGatewayTimeout,
		ETCDEndpointsStr: DefaultETCDEndpoints,
		ETCDUsername:     DefaultETCDUsername,
		ETCDPassword:     DefaultETCDPassword,
		ETCDTTL:          DefaultETCDTTL,
		RedisDSN:         DefaultRedisDSN,
		DBDSN:            DefaultDBDSN,
		Env:              DefaultEnv,
		ShutdownTimeout:  DefaultShutdownTimeout,
	}
}

func setDefaults(v *viper.Viper) {
	defaults := Defaults()
	v.SetDefault("http.addr", defaults.HTTPAddr)
	v.SetDefault("http.route_prefix", defaults.RoutePrefix)
	v.SetDefault("grpc.addr", defaults.GRPCAddr)
	v.SetDefault("grpc.service", defaults.ETCDServiceName)
	v.SetDefault("grpc.gateway_service", defaults.GatewayService)
	v.SetDefault("grpc.gateway_timeout", defaults.GatewayTimeout.String())
	v.SetDefault("etcd.endpoints", defaults.ETCDEndpointsStr)
	v.SetDefault("etcd.username", defaults.ETCDUsername)
	v.SetDefault("etcd.password", defaults.ETCDPassword)
	v.SetDefault("etcd.ttl", defaults.ETCDTTL.String())
	v.SetDefault("redis.dsn", defaults.RedisDSN)
	v.SetDefault("db.dsn", defaults.DBDSN)
	v.SetDefault("env", defaults.Env)
	v.SetDefault("shutdown.timeout", defaults.ShutdownTimeout.String())
}

func bindEnv(v *viper.Viper) {
	v.SetEnvPrefix("KIM")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()
	must(v.BindEnv("http.addr", "KIM_HTTP_ADDR"))
	must(v.BindEnv("http.route_prefix", "KIM_ROUTE_PREFIX"))
	must(v.BindEnv("grpc.addr", "KIM_GRPC_ADDR"))
	must(v.BindEnv("grpc.service", "KIM_GRPC_SERVICE"))
	must(v.BindEnv("grpc.gateway_service", "KIM_GATEWAY_SERVICE"))
	must(v.BindEnv("grpc.gateway_timeout", "KIM_GATEWAY_TIMEOUT"))
	must(v.BindEnv("etcd.endpoints", "KIM_ETCD_ENDPOINTS"))
	must(v.BindEnv("etcd.username", "KIM_ETCD_USERNAME"))
	must(v.BindEnv("etcd.password", "KIM_ETCD_PASSWORD"))
	must(v.BindEnv("etcd.ttl", "KIM_ETCD_TTL"))
	must(v.BindEnv("redis.dsn", "KIM_REDIS_DSN"))
	must(v.BindEnv("db.dsn", "KIM_DB_DSN"))
	must(v.BindEnv("env", "KIM_ENV"))
	must(v.BindEnv("shutdown.timeout", "KIM_SHUTDOWN_TIMEOUT"))
}

func bindFlags(v *viper.Viper, fs *pflag.FlagSet) error {
	bindings := map[string]string{
		"http.addr":             "http-addr",
		"http.route_prefix":     "route-prefix",
		"grpc.addr":             "grpc-addr",
		"grpc.service":          "grpc-service",
		"grpc.gateway_service":  "gateway-service",
		"grpc.gateway_timeout":  "gateway-timeout",
		"etcd.endpoints":        "etcd-endpoints",
		"etcd.username":         "etcd-username",
		"etcd.password":         "etcd-password",
		"etcd.ttl":              "etcd-ttl",
		"redis.dsn":             "redis-dsn",
		"db.dsn":                "db-dsn",
		"env":                   "env",
		"shutdown.timeout":      "shutdown-timeout",
	}
	for key, name := range bindings {
		if err := v.BindPFlag(key, fs.Lookup(name)); err != nil {
			return fmt.Errorf("bind flag %s: %w", name, err)
		}
	}
	return nil
}

func readConfigFile(v *viper.Viper, configPath string) error {
	configPath = strings.TrimSpace(configPath)
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yml")
		v.AddConfigPath(".")
	}
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok && configPath == "" {
			return nil
		}
		return fmt.Errorf("read config file: %w", err)
	}
	return nil
}

func (c *Config) normalize() {
	c.HTTPAddr = strings.TrimSpace(c.HTTPAddr)
	c.RoutePrefix = normalizePath(c.RoutePrefix)
	c.GRPCAddr = strings.TrimSpace(c.GRPCAddr)
	c.ETCDServiceName = strings.TrimSpace(c.ETCDServiceName)
	c.GatewayService = strings.TrimSpace(c.GatewayService)
	c.ETCDEndpointsStr = strings.TrimSpace(c.ETCDEndpointsStr)
	if c.ETCDEndpointsStr != "" {
		parts := strings.Split(c.ETCDEndpointsStr, ",")
		c.ETCDEndpoints = make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				c.ETCDEndpoints = append(c.ETCDEndpoints, p)
			}
		}
	}
	c.ETCDUsername = strings.TrimSpace(c.ETCDUsername)
	c.ETCDPassword = strings.TrimSpace(c.ETCDPassword)
	c.Env = strings.TrimSpace(c.Env)
	c.RedisDSN = strings.TrimSpace(c.RedisDSN)
	c.DBDSN = strings.TrimSpace(c.DBDSN)
}

// EtcdServiceName returns the service name with env suffix for etcd registration/discovery.
// For example, with service "kim" and env "dev", this returns "kim-dev".
func (c Config) EtcdServiceName() string {
	if c.Env == "" {
		return c.ETCDServiceName
	}
	return c.ETCDServiceName + "-" + c.Env
}

// GatewayServiceName returns the gateway service name with env suffix for etcd discovery.
func (c Config) GatewayServiceName() string {
	if c.Env == "" {
		return c.GatewayService
	}
	return c.GatewayService + "-" + c.Env
}

func (c Config) Validate() error {
	if c.HTTPAddr == "" {
		return errors.New("http addr is required")
	}
	if c.RoutePrefix == "" {
		return errors.New("route prefix is required")
	}
	if c.GRPCAddr == "" {
		return errors.New("grpc addr is required")
	}
	if c.ETCDServiceName == "" {
		return errors.New("etcd service name is required")
	}
	if c.GatewayService == "" {
		return errors.New("gateway service name is required")
	}
	if c.ShutdownTimeout <= 0 {
		return errors.New("shutdown timeout must be positive")
	}
	if c.RedisDSN == "" {
		return errors.New("redis dsn is required")
	}
	if c.DBDSN == "" {
		return errors.New("db dsn is required")
	}
	return nil
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func normalizeFlagArgs(args []string) []string {
	if len(args) == 0 {
		return nil
	}
	normalized := make([]string, len(args))
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && len(arg) > 2 {
			normalized[i] = "-" + arg
			continue
		}
		normalized[i] = arg
	}
	return normalized
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

package redis

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/go-redis/redis/v8"
)

type Redis interface {
	redis.Cmdable
	Close() error
}

type Config struct {
	Username  string
	Password  string
	Endpoints []string
	Dialer    func(ctx context.Context, network, addr string) (net.Conn, error)
	OnConnect func(ctx context.Context, cn *redis.Conn) error
	TlsConfig *tls.Config
}

type redisClient struct {
	Redis
}

func New(conf *Config) (Redis, error) {
	var con Redis
	var err error
	if len(conf.Endpoints) > 1 {
		con, err = loadCluster(conf)
		if err != nil {
			return nil, err
		}
	} else {
		con, err = loadClient(conf)
		if err != nil {
			return nil, err
		}
	}
	return &redisClient{
		Redis: con,
	}, nil
}

var _ Redis = (*redisClient)(nil)

func loadClient(c *Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:      c.Endpoints[0],
		Username:  c.Username,
		Password:  c.Password,
		Dialer:    c.Dialer,
		OnConnect: c.OnConnect,
		TLSConfig: c.TlsConfig,
	})
	err := client.Ping(context.Background()).Err()
	return client, err
}

func loadCluster(c *Config) (*redis.ClusterClient, error) {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:     c.Endpoints,
		Username:  c.Username,
		Password:  c.Password,
		Dialer:    c.Dialer,
		OnConnect: c.OnConnect,
		TLSConfig: c.TlsConfig,
	})
	err := client.Ping(context.Background()).Err()
	return client, err
}

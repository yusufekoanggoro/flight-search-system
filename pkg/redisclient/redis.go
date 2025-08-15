package redisclient

import (
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Addr     string
	Password string
	DB       int
}

type Client interface {
	Client() redis.UniversalClient
	Close() error
}

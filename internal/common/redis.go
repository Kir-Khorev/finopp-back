package common

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/Kir-Khorev/finopp-back/pkg/config"
)

func InitRedis(cfg *config.Config) *redis.Client {
	opts := &redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
	}

	// For production (Upstash) - use TLS and password
	if cfg.Environment == "production" {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		if cfg.RedisPassword != "" {
			opts.Password = cfg.RedisPassword
		}
	}

	rdb := redis.NewClient(opts)

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	log.Println("✅ Redis connected")
	return rdb
}


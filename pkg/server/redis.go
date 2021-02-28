package server

import (
	"github.com/go-redis/redis/v8"
)

//var RedisClient *redis.Client
//
//var onceRedisClient sync.Once
//
//func init() {
//
//	once.Do(func() {
//
//		RedisClient = GetRedisClient()
//	})
//}

func GetRedisClient(redisClusterAddr string) *redis.Client{
	return redis.NewClient(&redis.Options{
		Network:            "",
		Addr:               redisClusterAddr,
		Dialer:             nil,
		OnConnect:          nil,
		Username:           "",
		Password:           "",
		DB:                 0,
		MaxRetries:         0,
		MinRetryBackoff:    0,
		MaxRetryBackoff:    0,
		DialTimeout:        0,
		ReadTimeout:        0,
		WriteTimeout:       0,
		PoolSize:           0,
		MinIdleConns:       0,
		MaxConnAge:         0,
		PoolTimeout:        0,
		IdleTimeout:        0,
		IdleCheckFrequency: 0,
		TLSConfig:          nil,
		Limiter:            nil,
	})
}

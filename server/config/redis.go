package config

import "verve/utils"

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func GetRedisConfig() *RedisConfig {
	return &RedisConfig{
		Addr:     utils.GetEnv("REDIS_ADDR", "localhost:6379"),
		Password: utils.GetEnv("REDIS_PASSWORD", ""),
		DB:       0,
	}
}

package config

import (
	"os"
)

type Config struct {
	PostgresURL string
	RedisAddr   string
	QueueName   string
}

func LoadConfig() Config {
	return Config{
		PostgresURL: getEnv("POSTGRES_URL", "postgres://user:password@localhost:5432/jobs?sslmode=disable"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		QueueName:   getEnv("QUEUE_NAME", "jobs"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

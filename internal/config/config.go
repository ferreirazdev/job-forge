package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv             string
	WorkerCount        int
	RabbitMQURL        string
	RabbitMQQueue      string
	RabbitMQExchange   string
	RabbitMQRoutingKey string
}

func Load() (Config, error) {
	_ = godotenv.Load() // opcional: se .env não existir, segue
	workerCount, err := strconv.Atoi(getEnv("WORKER_COUNT", "4"))
	if err != nil {
		workerCount = 4
	}
	return Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		WorkerCount:        workerCount,
		RabbitMQURL:        os.Getenv("RABBITMQ_URL"),
		RabbitMQQueue:      os.Getenv("RABBITMQ_QUEUE"),
		RabbitMQExchange:   os.Getenv("RABBITMQ_EXCHANGE"),
		RabbitMQRoutingKey: os.Getenv("RABBITMQ_ROUTING_KEY"),
	}, nil
}
func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

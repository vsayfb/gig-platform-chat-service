package config

import (
	"log"
	"os"
)

type Config struct {
	ServiceName string
	AppEnv      string
	Port        string
	JWTSecret   string

	MongoURI string
	MongoDB  string

	AWSRegion   string
	SQSEndpoint string
	SQSQueueURL string

	UserServiceGRPCAddr string
	MetricsServerPort   string
	OTelCollectorAddr   string
}

func Load() *Config {
	return &Config{
		ServiceName:         mustGetEnv("SERVICE_NAME"),
		AppEnv:              mustGetEnv("APP_ENV"),
		Port:                mustGetEnv("SERVER_PORT"),
		JWTSecret:           mustGetEnv("JWT_SECRET"),
		MongoURI:            mustGetEnv("MONGO_URI"),
		MongoDB:             mustGetEnv("MONGO_DB"),
		AWSRegion:           mustGetEnv("AWS_REGION"),
		SQSEndpoint:         os.Getenv("SQS_ENDPOINT"),
		SQSQueueURL:         mustGetEnv("SQS_QUEUE_URL"),
		UserServiceGRPCAddr: mustGetEnv("USER_SERVICE_GRPC_ADDR"),
		MetricsServerPort:   mustGetEnv("METRICS_SERVER_PORT"),
		OTelCollectorAddr:   mustGetEnv("OTEL_COLLECTOR_ADDR"),
	}
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env var: %s", key)
	}
	return v
}

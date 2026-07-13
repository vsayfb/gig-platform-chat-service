package config

import (
	"log"
	"os"
)

type Config struct {
	ServiceName string
	AppEnv      string
	Port        string
	GRPCPort    string
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
		AppEnv:              getEnv("APP_ENV", "production"),
		Port:                getEnv("REST_PORT", "8080"),
		GRPCPort:            getEnv("GRPC_PORT", "9090"),
		JWTSecret:           mustGetEnv("JWT_SECRET"),
		MongoURI:            mustGetEnv("MONGO_URI"),
		MongoDB:             mustGetEnv("MONGO_DB"),
		AWSRegion:           mustGetEnv("AWS_REGION"),
		SQSEndpoint:         os.Getenv("SQS_ENDPOINT"),
		SQSQueueURL:         mustGetEnv("SQS_QUEUE_URL"),
		UserServiceGRPCAddr: mustGetEnv("USER_SERVICE_GRPC_ADDR"),
		MetricsServerPort:   getEnv("METRICS_SERVER_PORT", ":9100"),
		OTelCollectorAddr:   getEnv("OTEL_COLLECTOR_ADDR", "localhost:4317"),
	}
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env var: %s", key)
	}
	return v
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

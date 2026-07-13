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

	UserServiceGRPCAddr string
	MetricsServerPort   string
	OTelCollectorAddr   string
}

func Load() *Config {
	return &Config{
		ServiceName:         mustGetEnv("SERVICE_NAME"),
		AppEnv:              getEnv("APP_ENV", "production"),
		Port:                getEnv("REST_PORT", "8081"),
		GRPCPort:            getEnv("GRPC_PORT", "9090"),
		JWTSecret:           mustGetEnv("JWT_SECRET"),
		MongoURI:            mustGetEnv("MONGO_URI"),
		MongoDB:             mustGetEnv("MONGO_DB"),
		UserServiceGRPCAddr: getEnv("USER_SERVICE_GRPC_ADDR", "localhost:8080"),
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

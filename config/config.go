package config

import (
	"log"
	"os"
)

type Config struct {
	AppEnv      string
	Port        string
	JWTSecret   string
	MongoURI    string
	MongoDB     string
	AWSRegion   string
	SQSQueueURL string
}

func Load() *Config {
	return &Config{
		AppEnv:      mustGetEnv("APP_ENV"),
		Port:        mustGetEnv("SERVER_PORT"),
		JWTSecret:   mustGetEnv("JWT_SECRET"),
		MongoURI:    mustGetEnv("MONGO_URI"),
		MongoDB:     mustGetEnv("MONGO_DB"),
		AWSRegion:   mustGetEnv("AWS_REGION"),
		SQSQueueURL: mustGetEnv("SQS_QUEUE_URL"),
	}
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env var: %s", key)
	}
	return v
}

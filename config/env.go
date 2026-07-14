package config

func loadEnv() (*Config, error) {
	return &Config{
		ServiceName:         getEnv("SERVICE_NAME", "chat-service"),
		AppEnv:              getEnv("APP_ENV", "development"),
		Port:                getEnv("SERVER_PORT", "8081"),
		GRPCPort:            getEnv("GRPC_PORT", "9090"),
		JWTSecret:           mustGetEnv("JWT_SECRET"),
		MongoURI:            mustGetEnv("MONGO_URI"),
		MongoDB:             mustGetEnv("MONGO_DB"),
		UserServiceGRPCAddr: getEnv("USER_SERVICE_GRPC_ADDR", "localhost:9090"),
		MetricsServerPort:   getEnv("METRICS_SERVER_PORT", ":9100"),
		OTelCollectorAddr:   getEnv("OTEL_COLLECTOR_ADDR", "localhost:4317"),
	}, nil
}

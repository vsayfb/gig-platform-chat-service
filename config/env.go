package config

func loadEnv() (*Config, error) {
	return &Config{
		ServiceName:         getEnv(EnvServiceName, DefaultServiceName),
		AppEnv:              EnvironmentDevelopment,
		Port:                getEnv(EnvServerPort, DefaultServerPort),
		JWTSecret:           mustGetEnv(EnvJWTSecret),
		MongoURI:            mustGetEnv(EnvMongoURI),
		MongoDB:             mustGetEnv(EnvMongoDB),
		UserServiceGRPCAddr: getEnv(EnvUserServiceGRPCAddr, DefaultUserServiceGRPCAddr),
		MetricsServerPort:   getEnv(EnvMetricsServerPort, DefaultMetricsServerPort),
		OTelCollectorAddr:   getEnv(EnvOTelCollectorAddr, DefaultOtelCollectorAddr),
	}, nil
}

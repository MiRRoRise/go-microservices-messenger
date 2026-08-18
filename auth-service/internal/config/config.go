package config

import "os"

type Config struct {
	DBUser     string
	DBPassword string
	DBName     string
	DBHost     string
	DBPort     string

	JWTSecret string

	ServerPort   string
	GRPCPort     string
	KafkaBrokers string
}

func New() *Config {
	return &Config{
		DBUser:     getEnv("DB_USER", "user"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "db"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),

		JWTSecret: getEnv("JWT_SECRET", "secret-key"),

		ServerPort:   getEnv("SERVER_PORT", ":8080"),
		GRPCPort:     getEnv("GRPC_PORT", ":50051"),
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

package config

import (
	"os"

	"github.com/alissoncorsair/appsolidario-backend/types"
)

var Envs = initConfig()

func initConfig() types.Config {
	return types.Config{
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresUser:     getEnv("POSTGRES_USER", "postgres"),
		PostgresDB:       getEnv("POSTGRES_DB", "postgres"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

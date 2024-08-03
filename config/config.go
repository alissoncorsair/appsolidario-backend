package config

import (
	"log"
	"os"
	"strconv"

	"github.com/alissoncorsair/appsolidario-backend/types"
	"github.com/joho/godotenv"
)

var Envs = initConfig()

func initConfig() *types.Config {
	godotenv.Load()

	port, err := strconv.Atoi(getEnv("POSTGRES_PORT", "5432"))

	if err != nil {
		log.Fatal(err)
	}

	return &types.Config{
		Host:                          getEnv("POSTGRES_HOST", "localhost"),
		Port:                          port,
		PostgresPassword:              getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresUser:                  getEnv("POSTGRES_USER", "postgres"),
		PostgresDB:                    getEnv("POSTGRES_DB", "postgres"),
		SSLMode:                       getEnv("POSTGRES_SSL_MODE", "disable"),
		JWTExpirationInSeconds:        getEnvAsInt64("JWT_EXPIRATION_IN_SECONDS", 3600),
		JWTSecret:                     getEnv("JWT_SECRET", "secret"),
		JWTRefreshExpirationInSeconds: getEnvAsInt64("JWT_REFRESH_EXPIRATION_IN_SECONDS", 60),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func getEnvAsInt64(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fallback
		}
		return i
	}

	return fallback
}

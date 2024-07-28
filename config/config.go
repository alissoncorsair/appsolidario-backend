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
		Host:             getEnv("POSTGRES_HOST", "database"),
		Port:             port,
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresUser:     getEnv("POSTGRES_USER", "postgres"),
		PostgresDB:       getEnv("POSTGRES_DB", "postgres"),
		SSLMode:          getEnv("POSTGRES_SSL_MODE", "disable"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

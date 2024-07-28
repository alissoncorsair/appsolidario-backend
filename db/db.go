package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/alissoncorsair/appsolidario-backend/types"
	_ "github.com/lib/pq"
)

func NewPostgreSQLStorage(config types.Config) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.PostgresUser, config.PostgresPassword, config.PostgresDB, config.SSLMode,
	)

	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		log.Fatal(err)
	}

	return db, nil
}

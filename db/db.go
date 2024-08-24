package db

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/alissoncorsair/appsolidario-backend/types"
	_ "github.com/lib/pq"
)

func NewPostgreSQLStorage(config types.Config) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.PostgresUser,
		config.PostgresPassword,
		config.Host,
		config.Port,
		config.PostgresDB,
		config.SSLMode,
	)

	// Add SSL root cert if provided
	if config.PGCert != "" {
		psqlInfo += fmt.Sprintf(";sslrootcert=%s", url.QueryEscape(config.PGCert))
	}

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	return db, nil
}

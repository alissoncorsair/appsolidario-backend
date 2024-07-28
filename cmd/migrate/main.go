package main

import (
	"log"
	"os"

	"github.com/alissoncorsair/appsolidario-backend/config"
	"github.com/alissoncorsair/appsolidario-backend/db"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	dbCfg := config.Envs
	db, err := db.NewPostgreSQLStorage(*dbCfg)

	if err != nil {
		log.Fatal(err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})

	if err != nil {
		log.Fatal("Error creating driver: ", err)
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://cmd/migrate/migrations", "postgres", driver)

	if err != nil {
		log.Fatal("Error creating migration: ", err)
		log.Fatal(err)
	}

	cmd := os.Args[len(os.Args)-1]

	switch cmd {
	case "up":
		err = m.Up()
		if err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
	case "down":
		err = m.Down()
		if err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
	default:
		log.Fatal("Invalid command")
	}
}

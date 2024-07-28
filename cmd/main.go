package main

import (
	"database/sql"
	"log"

	"github.com/alissoncorsair/appsolidario-backend/cmd/api"
	"github.com/alissoncorsair/appsolidario-backend/config"
	"github.com/alissoncorsair/appsolidario-backend/db"
)

func main() {
	cfg := config.Envs
	db, err := db.NewPostgreSQLStorage(*cfg)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	initStorage(db)

	server := api.NewAPIServer(":8080", db)
	server.Run()
}

func initStorage(db *sql.DB) {
	err := db.Ping()

	if err != nil {
		log.Fatal(err)
	}

	log.Println("connected to db")
}

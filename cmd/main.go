package main

import "github.com/alissoncorsair/appsolidario-backend/cmd/api"

func main() {
	server := api.NewAPIServer(":8080")
	server.Run()
}

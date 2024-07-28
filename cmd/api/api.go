package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/alissoncorsair/appsolidario-backend/utils"
)

type APIServer struct {
	addr string
	db   *sql.DB
}

func NewAPIServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
	}
}

func (s *APIServer) Run() error {
	router := http.NewServeMux()
	router.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Hello, World!!"})
	})

	server := http.Server{
		Addr:    s.addr,
		Handler: router,
	}

	log.Printf("Server has started %s", s.addr)

	return server.ListenAndServe()
}

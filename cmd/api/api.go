package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/alissoncorsair/appsolidario-backend/config"
	"github.com/alissoncorsair/appsolidario-backend/service/mailer"
	"github.com/alissoncorsair/appsolidario-backend/service/post"
	"github.com/alissoncorsair/appsolidario-backend/service/user"
	"github.com/alissoncorsair/appsolidario-backend/storage"
	"github.com/alissoncorsair/appsolidario-backend/utils"
)

type APIServer struct {
	addr    string
	db      *sql.DB
	storage *storage.R2Storage
}

func NewAPIServer(addr string, db *sql.DB) *APIServer {
	accountID := config.Envs.R2AccountID
	bucketName := config.Envs.R2BucketName
	storage, err := storage.NewR2Storage(accountID, bucketName)

	if err != nil {
		log.Fatalf("Failed to create R2 storage: %v", err)
	}

	return &APIServer{
		addr:    addr,
		db:      db,
		storage: storage,
	}
}

func (s *APIServer) Run() error {
	router := http.NewServeMux()

	apiRouter := http.NewServeMux()

	apiRouter.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Hello, World!!"})
	})

	mailer := mailer.NewSendGridMailer(config.Envs.SendgridApiKey, config.Envs.EmailFrom, config.Envs.DevMode)
	userStore := user.NewStore(s.db)
	userHandler := user.NewHandler(userStore, mailer)
	userHandler.RegisterRoutes(apiRouter)
	postStore := post.NewStore(s.db)
	postHandler := post.NewHandler(postStore, userStore, s.storage)
	postHandler.RegisterRoutes(apiRouter)

	router.Handle("/api/", corsMiddleware(http.StripPrefix("/api", apiRouter)))

	server := http.Server{
		Addr:    s.addr,
		Handler: router,
	}

	log.Printf("Server has started %s", s.addr)

	return server.ListenAndServe()
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

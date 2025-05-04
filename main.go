package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/savisitor15/go-http-serv/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbConnection   *database.Queries
}

func main() {
	godotenv.Load(".env")
	dbUrl := os.Getenv("DB_URL")
	const filepathRoot = "."
	const port = "8080"

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal(err)
	}
	// Common STATE
	var apiCfg apiConfig = apiConfig{fileserverHits: atomic.Int32{}}
	// asign the db connector
	apiCfg.dbConnection = database.New(db)

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	// API
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	// Chirp management
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerAddChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	// Users management
	mux.HandleFunc("POST /api/users", apiCfg.handleUserCreation)

	// Admin
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

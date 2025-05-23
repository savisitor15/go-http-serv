package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/savisitor15/go-http-serv/internal/auth"
	"github.com/savisitor15/go-http-serv/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbConnection   *database.Queries
	supserSecret   string
	polkaSecret    string
}

func main() {
	godotenv.Load(".env")
	dbUrl := os.Getenv("DB_URL")
	secret := os.Getenv("SECRET")
	polkaAPIKey := os.Getenv("POLKA_KEY")
	const filepathRoot = "."
	const port = "8080"

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal(err)
	}
	// Common STATE
	var apiCfg apiConfig = apiConfig{fileserverHits: atomic.Int32{}}
	// assign token generator secret
	apiCfg.supserSecret = secret
	// assign the api key for polka
	apiCfg.polkaSecret = polkaAPIKey
	// asign the db connector
	apiCfg.dbConnection = database.New(db)

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	// API
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	// Chirp management
	mux.Handle("POST /api/chirps", apiCfg.middlewareAuthenticated(http.HandlerFunc(apiCfg.handlerAddChirp)))
	mux.Handle("DELETE /api/chirps/{chirpID}", apiCfg.middlewareAuthenticated(http.HandlerFunc(apiCfg.handlerDeleteChirp)))
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirpByID)
	// Users management
	mux.HandleFunc("POST /api/users", apiCfg.handleUserCreation)
	mux.HandleFunc("POST /api/login", apiCfg.handlerCheckLogin)
	mux.Handle("PUT /api/users", apiCfg.middlewareAuthenticated(http.HandlerFunc(apiCfg.handlerUpdatePassword)))
	mux.Handle("POST /api/refresh", apiCfg.middlewareRefToken(http.HandlerFunc(apiCfg.handlerRefreshToken)))
	mux.Handle("POST /api/revoke", apiCfg.middlewareRefToken(http.HandlerFunc(apiCfg.handlerRevokeToken)))
	// Webhooks
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerProcessChirpyRed)
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

func (cfg *apiConfig) middlewareAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := cfg.GetUserAuthority(r)
		if err != nil {
			log.Println("middlewareAuthenticated() failed", err)
			errorJSONBody(w, 401, err)
		}
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) middlewareRefToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tok, err := auth.GetBearerToken(r.Header)
		if err != nil {
			log.Println("Unauthorized access! RefreshToken", err)
			errorJSONBody(w, 401, errors.New("Unauthorized"))
			return
		}
		_, err = cfg.FindRefreshToken(r, tok)
		if err != nil {
			log.Println("Unauthorized access! RefreshToken", err)
			errorJSONBody(w, 401, errors.New("Unauthorized"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

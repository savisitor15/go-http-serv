package main

import (
	"log"
	"net/http"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	err := cfg.dbConnection.DestroyAllUsers(r.Context())
	if err != nil {
		log.Println("failed to clear users", err)
	} else {
		log.Println("all users deleted!")
	}
	err = cfg.dbConnection.DestroyAllchirps(r.Context())
	if err != nil {
		log.Println("failed to clear chirps", err)
	} else {
		log.Println("all chirps deleted!")
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}

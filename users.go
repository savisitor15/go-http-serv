package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/savisitor15/go-http-serv/internal/database"
)

func (cfg *apiConfig) handleUserCreation(w http.ResponseWriter, r *http.Request) {
	type reqIn struct {
		Email string `json:"email"`
	}
	type resOut struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}
	reqin := reqIn{}
	err := decodeRequestBody(r, &reqin)
	if err != nil {
		log.Println("unable to parse user post body", err)
		errorJSONBody(w, 500, err)
		return
	}
	ts := time.Now()
	params := database.CreateUserParams{CreatedAt: ts, UpdatedAt: ts, Email: sql.NullString{Valid: true, String: reqin.Email}}
	dbuser, err := cfg.dbConnection.CreateUser(r.Context(), params)
	if err != nil {
		log.Println("user creation failed!", err)
		errorJSONBody(w, 500, err)
		return
	}
	respondJSONBody(w, 201, resOut{
		ID:        dbuser.ID,
		CreatedAt: dbuser.CreatedAt,
		UpdatedAt: dbuser.UpdatedAt,
		Email:     dbuser.Email.String,
	})
}

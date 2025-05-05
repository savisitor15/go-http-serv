package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/savisitor15/go-http-serv/internal/auth"
	"github.com/savisitor15/go-http-serv/internal/database"
)

type UserJSON struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func convertDbUserToJSON(u database.User) UserJSON {
	return UserJSON{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Email:     u.Email,
	}
}

func (cfg *apiConfig) handleUserCreation(w http.ResponseWriter, r *http.Request) {
	type reqIn struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	reqin := reqIn{}
	err := decodeRequestBody(r, &reqin)
	if err != nil {
		log.Println("unable to parse user post body", err)
		errorJSONBody(w, 500, err)
		return
	}
	if len(reqin.Email) == 0 || len(reqin.Password) == 0 {
		log.Println("invalid user request!")
		errorJSONBody(w, 400, errors.New("Missing email or password"))
		return
	}
	ts := time.Now()
	hashPass, err := auth.HashPassword(reqin.Password)
	if err != nil {
		log.Println("CreateUser password hashing failed", err)
		errorJSONBody(w, 500, err)
		return
	}
	params := database.CreateUserParams{CreatedAt: ts, UpdatedAt: ts, Email: reqin.Email, HashedPassword: hashPass}
	dbuser, err := cfg.dbConnection.CreateUser(r.Context(), params)
	if err != nil {
		log.Println("user creation failed!", err)
		errorJSONBody(w, 500, err)
		return
	}
	respondJSONBody(w, 201, convertDbUserToJSON(dbuser))
}

func (cfg *apiConfig) handlerCheckLogin(w http.ResponseWriter, r *http.Request) {
	type reqIn struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	// Out is a UserJSON
	reqin := reqIn{}
	err := decodeRequestBody(r, &reqin)
	if err != nil {
		log.Println("unable to parse user post body", err)
		errorJSONBody(w, 500, err)
		return
	}
	dbuser, err := cfg.dbConnection.GetUserByEmail(r.Context(), reqin.Email)
	if err != nil {
		errorJSONBody(w, 404, errors.New("User not found"))
		return
	}
	// check password
	err = auth.CheckPasswordHash(dbuser.HashedPassword, reqin.Password)
	if err != nil {
		errorJSONBody(w, 401, errors.New("Unauthorized access"))
		return
	}
	respondJSONBody(w, 200, convertDbUserToJSON(dbuser))
}

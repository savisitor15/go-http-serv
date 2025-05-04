package main

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/savisitor15/go-http-serv/internal/database"
)

func censorProfanity(s string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	var out []string
	bad := false
	for _, elm := range strings.Split(s, " ") {
		for _, badWord := range badWords {
			if strings.ToLower(elm) == badWord {
				bad = true
				break
			}
		}
		if bad {
			out = append(out, "****")
			bad = false
		} else {
			out = append(out, elm)
		}
	}
	return strings.Join(out, " ")
}

func (cfg *apiConfig) handlerAddChirp(w http.ResponseWriter, r *http.Request) {
	type reqIn struct {
		Body string    `json:"body"`
		User uuid.UUID `json:"user_id"`
	}
	type resOut struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		User      uuid.UUID `json:"user_id"`
	}
	reqin := reqIn{}
	err := decodeRequestBody(r, &reqin)
	if err != nil {
		log.Println(err)
		errorJSONBody(w, 500, err)
		return
	}
	if len(reqin.Body) > 140 {
		log.Println("Chirp is too long!")
		errorJSONBody(w, 400, errors.New("Chirp is too long"))
		return
	}
	// We should have a valid chirp to add now
	ts := time.Now()
	params := database.CreateChirpParams{
		CreatedAt: ts,
		UpdatedAt: ts,
		Body:      censorProfanity(reqin.Body),
		UserID:    reqin.User,
	}
	chirp, err := cfg.dbConnection.CreateChirp(r.Context(), params)
	if err != nil {
		log.Println("Error creating chirp", err)
		errorJSONBody(w, 500, err)
	}
	out := resOut{}
	out.ID = chirp.ID
	out.Body = chirp.Body
	out.CreatedAt = chirp.CreatedAt
	out.UpdatedAt = chirp.UpdatedAt
	out.User = chirp.UserID
	respondJSONBody(w, 201, out)
}

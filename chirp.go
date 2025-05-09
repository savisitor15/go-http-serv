package main

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/savisitor15/go-http-serv/internal/auth"
	"github.com/savisitor15/go-http-serv/internal/database"
)

type ChirpJSON struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (c *ChirpJSON) ConvertChirpFromDB(chirp database.Chirp) ChirpJSON {
	c.ID = chirp.ID
	c.Body = chirp.Body
	c.CreatedAt = chirp.CreatedAt
	c.UpdatedAt = chirp.UpdatedAt
	c.UserID = chirp.UserID
	return *c
}

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

func ChirpIDFromPath(r *http.Request) (uuid.UUID, error) {
	chirpid := r.PathValue("chirpID")
	if len(chirpid) == 0 {
		return uuid.Nil, errors.New("chirpID not in Path")
	}
	uid, err := uuid.Parse(chirpid)
	if err != nil {
		return uuid.Nil, err
	}
	return uid, nil
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	uid, err := ChirpIDFromPath(r)
	if err != nil {
		log.Println("handlerDeleteChirp() failed to parse chirp id", err)
		errorJSONBody(w, 500, errors.New("error parsing chirpid"))
		return
	}
	user, err := cfg.GetUserAuthority(r)
	if err != nil {
		errorJSONBody(w, 401, err)
		return
	}
	chirp, err := cfg.dbConnection.GetChirpByID(r.Context(), uid)
	if err != nil {
		log.Println("handlerDeleteChirp() failed to find the chirp", err)
		errorJSONBody(w, 404, errors.New("failed to find chirp"))
		return
	}
	if chirp.UserID != user.ID {
		// NOT ALLOWED
		log.Println("handlerDeleteChirp() user trying to delete someone else's chirp")
		errorJSONBody(w, 403, nil)
		return
	}
	err = cfg.dbConnection.DeleteChirpByID(r.Context(), chirp.ID)
	if err != nil {
		log.Println("handlerDeleteChirp() error deleting chirp from data table", err)
		errorJSONBody(w, 500, errors.New("server side error"))
		return
	}
	respondJSONBody(w, 204, nil)
}

func (cfg *apiConfig) handlerAddChirp(w http.ResponseWriter, r *http.Request) {
	type reqIn struct {
		Body string    `json:"body"`
		User uuid.UUID `json:"user_id"`
	}
	reqin := reqIn{}
	err := decodeRequestBody(r, &reqin)
	if err != nil {
		log.Println(err)
		errorJSONBody(w, 500, err)
		return
	}
	if reqin.User == uuid.Nil {
		tok, err := auth.GetBearerToken(r.Header)
		if err != nil {
			log.Println(err)
			errorJSONBody(w, 500, err)
			return
		}
		uid, err := auth.ValidateJWT(tok, cfg.supserSecret)
		if err != nil {
			log.Println(err)
			errorJSONBody(w, 500, err)
			return
		}
		reqin.User = uid
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
	out := ChirpJSON{}
	out.ConvertChirpFromDB(chirp)
	respondJSONBody(w, 201, out)
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.dbConnection.GetAllChirps(r.Context())
	if err != nil {
		log.Println("Error getting chirps", err)
	}
	// reformat so JSON stays consistent
	out := make([]ChirpJSON, len(chirps))
	for idx, elm := range chirps {
		out[idx] = ChirpJSON(elm)
	}
	respondJSONBody(w, 200, out)
}

func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	uid, err := ChirpIDFromPath(r)
	if err != nil {
		errorJSONBody(w, 500, err)
		return
	}
	chirp, err := cfg.dbConnection.GetChirpByID(r.Context(), uid)
	if err != nil {
		errorJSONBody(w, 404, err)
		return
	}
	out := ChirpJSON{}
	out.ConvertChirpFromDB(chirp)
	respondJSONBody(w, 200, out)
}

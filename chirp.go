package main

import (
	"errors"
	"log"
	"net/http"
	"strings"
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

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type reqIn struct {
		Body string `json:"body"`
	}
	type resOut struct {
		Valid       bool   `json:"valid"`
		CleanedBody string `json:"cleaned_body"`
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
	out := resOut{Valid: true}
	out.CleanedBody = censorProfanity(reqin.Body)
	respondJSONBody(w, 200, out)
}

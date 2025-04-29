package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

func errorBody(er error) []byte {
	type errorOut struct {
		Error string `json:"error"`
	}
	out := errorOut{
		Error: er.Error(),
	}
	dat, err := json.Marshal(out)
	if err != nil {
		return []byte(err.Error())
	}
	return dat
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type reqIn struct {
		Body string `json:"body"`
	}
	type resOut struct {
		Valid bool `json:"valid"`
	}
	decoder := json.NewDecoder(r.Body)
	reqin := reqIn{}
	err := decoder.Decode(&reqin)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write(errorBody(err))
		return
	}
	if len(reqin.Body) > 140 {
		log.Println("Chirp is too long!")
		w.WriteHeader(400)
		w.Write(errorBody(errors.New("Chirp is too long")))
		return
	}
	dat, err := json.Marshal(resOut{Valid: true})
	if err != nil {
		w.WriteHeader(500)
		w.Write(errorBody(err))
		return
	}
	w.WriteHeader(200)
	w.Write(dat)
}

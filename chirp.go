package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

func errorJSONBody(w http.ResponseWriter, returnCode int, er error) {
	w.Header().Set("Content-Type", "application/json")
	if returnCode <= 0 {
		returnCode = 500 // Default
	}
	type errorOut struct {
		Error string `json:"error"`
	}
	out := errorOut{
		Error: er.Error(),
	}
	dat, err := json.Marshal(out)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	w.WriteHeader(returnCode)
	w.Write(dat)
	return
}

func respondJSONBody(w http.ResponseWriter, returnCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err == nil {
		w.WriteHeader(returnCode)
		_, err := w.Write([]byte(dat))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		errorJSONBody(w, 500, err)
		log.Fatal(err)
	}

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
	respondJSONBody(w, 200, resOut{Valid: true})
}

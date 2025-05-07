package main

import (
	"encoding/json"
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

func respondJSONBody(w http.ResponseWriter, returnCode int, payload any) {
	if payload == nil {
		w.WriteHeader(returnCode)
		return
	}
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

func decodeRequestBody(r *http.Request, payload any) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(&payload)
}

package main

import (
	"log"
	"net/http"

	"github.com/google/uuid"
)

type PolkaRequest struct {
	Event string `json:"event"`
	Data  struct {
		UserID string `json:"user_id"`
	} `json:"data"`
}

func (cfg apiConfig) handlerProcessChirpyRed(w http.ResponseWriter, r *http.Request) {
	inreq := PolkaRequest{}
	err := decodeRequestBody(r, &inreq)
	if err != nil {
		log.Println("handlerProcessChirpyRed(), failed to decode request body", err)
		errorJSONBody(w, 500, nil)
		return
	}
	if inreq.Event != "user.upgraded" {
		log.Println("handlerProcessChirpyRed() unwanted event", inreq.Event)
		respondJSONBody(w, 204, nil)
		return
	}
	uid, err := uuid.Parse(inreq.Data.UserID)
	if err != nil {
		log.Println("handlerProcessChirpyRed() error decoding the uid", err)
		errorJSONBody(w, 404, nil)
		return
	}
	_, err = cfg.dbConnection.SetChirpyRedByID(r.Context(), uid)
	if err != nil {
		log.Println("handlerProcessChirpyRed() error decoding the uid", err)
		errorJSONBody(w, 404, nil)
		return
	}
	respondJSONBody(w, 204, nil)
}

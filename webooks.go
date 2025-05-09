package main

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/savisitor15/go-http-serv/internal/auth"
)

type PolkaRequest struct {
	Event string `json:"event"`
	Data  struct {
		UserID string `json:"user_id"`
	} `json:"data"`
}

func (cfg *apiConfig) handlerProcessChirpyRed(w http.ResponseWriter, r *http.Request) {
	inreq := PolkaRequest{}
	key, err := auth.GetAPIKey(r.Header)
	if err != nil {
		log.Println("handlerProcessChirpyRed(), failed to get api key", err)
		errorJSONBody(w, 401, nil)
		return
	}
	if key != cfg.polkaSecret {
		log.Println("handlerProcessChirpyRed(), unauthorized access! given:", key, "expected:", cfg.polkaSecret)
		errorJSONBody(w, 401, nil)
		return
	}
	err = decodeRequestBody(r, &inreq)
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

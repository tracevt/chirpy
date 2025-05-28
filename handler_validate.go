package main

import (
	"encoding/json"
	"net/http"
)

func handleChirpsValidate(w http.ResponseWriter, r *http.Request) {
	type message struct {
		Body string `json:"body"`
	}

	type validMessage struct {
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(r.Body)
	params := message{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	respondWithJSON(w, http.StatusOK, validMessage{
		Valid: true,
	})
}

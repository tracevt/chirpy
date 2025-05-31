package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func containsBadWords(s string) bool {
	if s == "kerfuffle" {
		return true
	} else if s == "sharbert" {
		return true
	} else if s == "fornax" {
		return true
	}

	return false
}

func handleChirpsValidate(w http.ResponseWriter, r *http.Request) {
	type message struct {
		Body string `json:"body"`
	}

	type cleanMessage struct {
		CleanedBody string `json:"cleaned_body"`
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

	splitMsg := strings.Split(params.Body, " ")

	for idx, component := range splitMsg {
		if containsBadWords(strings.ToLower(component)) {
			splitMsg[idx] = "****"
		}
	}

	joinMsg := strings.Join(splitMsg, " ")
	respondWithJSON(w, http.StatusOK, cleanMessage{
		CleanedBody: joinMsg,
	})
}

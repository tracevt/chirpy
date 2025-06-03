package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tracevt/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

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

func (cfg *apiConfig) handleChirps(w http.ResponseWriter, r *http.Request) {
	type message struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
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

	if params.UserID.String() == "" {
		respondWithError(w, http.StatusBadRequest, "Please provide the user that will post the chirp", nil)
		return
	}

	splitMsg := strings.Split(params.Body, " ")

	for idx, component := range splitMsg {
		if containsBadWords(strings.ToLower(component)) {
			splitMsg[idx] = "****"
		}
	}

	joinMsg := strings.Join(splitMsg, " ")

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   joinMsg,
		UserID: params.UserID,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
		return
	}

	jsonChirp := &Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	respondWithJSON(w, http.StatusCreated, jsonChirp)
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetChirps(r.Context())

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps", err)
		return
	}

	jsonChirps := make([]Chirp, 0)
	for _, chirp := range chirps {
		t := &Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}

		jsonChirps = append(jsonChirps, *t)
	}

	respondWithJSON(w, http.StatusOK, jsonChirps)
}

func (cfg *apiConfig) getChirp(w http.ResponseWriter, r *http.Request) {
	chirpId := r.PathValue("chirpID")

	if chirpId == "" {
		respondWithError(w, http.StatusBadRequest, "Please provide a ChirpID", fmt.Errorf("ChirpID not provided"))
		return
	}

	chirpUUID, err := uuid.Parse(chirpId)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Chirp UUID is not in the correct format", err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpUUID)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	jsonChirp := &Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	respondWithJSON(w, http.StatusOK, jsonChirp)
}

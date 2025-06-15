package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tracevt/chirpy/internal/auth"
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
		Body string `json:"body"`
	}

	// Validate if the Token is valid
	bearerToken, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Please provide an auth token", err)
		return
	}

	userIDFromToken, err := auth.ValidateJWT(bearerToken, cfg.secret)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate token", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := message{}
	err = decoder.Decode(&params)
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

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   joinMsg,
		UserID: userIDFromToken,
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
	// Check for author_id
	author := r.URL.Query().Get("author_id")

	// sort direction
	sortDirection := "asc"
	sortDirectionParam := r.URL.Query().Get("sort")
	if sortDirectionParam == "desc" {
		sortDirection = "desc"
	}

	if author != "" {
		authorUUID, err := uuid.Parse(author)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Author ID in the wrong format", err)
			return
		}

		chirps, err := cfg.db.GetChirpsByAuthor(r.Context(), authorUUID)

		parseChirps(chirps, err, w, sortDirection)
	} else {
		chirps, err := cfg.db.GetChirps(r.Context())

		parseChirps(chirps, err, w, sortDirection)
	}
}

func parseChirps(chirps []database.Chirp, err error, w http.ResponseWriter, sortDirection string) {
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

	sort.Slice(jsonChirps, func(i, j int) bool {
		if sortDirection == "desc" {
			return jsonChirps[i].CreatedAt.After(jsonChirps[j].CreatedAt)
		}
		return jsonChirps[i].CreatedAt.Before(jsonChirps[j].CreatedAt)
	})

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

func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {
	// Check for the token in the headers
	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Please provide an auth token", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.secret)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate token", err)
		return
	}

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

	if userID != chirp.UserID {
		respondWithError(w, http.StatusForbidden, "You're not allowed to delete that", err)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), chirpUUID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete the chirp", err)
		return
	}

	respondWithNoContent(w)
}

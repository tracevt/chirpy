package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tracevt/chirpy/internal/auth"
)

func (cfg *apiConfig) login(w http.ResponseWriter, r *http.Request) {
	type loginCredentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := loginCredentials{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if params.Email == "" {
		respondWithError(w, http.StatusBadRequest, "Email is required", fmt.Errorf("Email is required"))
		return
	}

	if params.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Password is required", fmt.Errorf("Email is required"))
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "incorrect email or password", err)
		return
	}

	noMatch := auth.CheckPasswordHash(user.HashedPassword, params.Password)

	if noMatch != nil {
		respondWithError(w, http.StatusUnauthorized, "incorrect email or password", noMatch)
		return
	}

	jsonUser := &User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(w, http.StatusOK, jsonUser)
}

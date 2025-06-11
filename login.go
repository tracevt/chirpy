package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/tracevt/chirpy/internal/auth"
	"github.com/tracevt/chirpy/internal/database"
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
		respondWithError(w, http.StatusBadRequest, "Email is required", fmt.Errorf("email is required"))
		return
	}

	if params.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Password is required", fmt.Errorf("password is required"))
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "incorrect email or password", err)
		return
	}

	noMatch := auth.CheckPasswordHash(params.Password, user.HashedPassword)

	if noMatch != nil {
		respondWithError(w, http.StatusUnauthorized, "incorrect email or password", noMatch)
		return
	}

	expiration := time.Minute * 60

	token, err := auth.MakeJWT(user.ID, cfg.secret, expiration)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "JWT couldn't be generated", err)
		return
	}

	// Create Refresh Token on the DB
	refreshToken := auth.MakeRefreshToken()
	refreshTokenExpiration := time.Minute * 86400 // 60 days after the fact
	refreshTokenExpirationDate := time.Now().Add(refreshTokenExpiration)

	refreshTokenDB, err := cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: refreshTokenExpirationDate,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Refresh Token couldn't be generated", err)
		return
	}

	jsonUser := &UserWithToken{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshTokenDB.Token,
	}

	respondWithJSON(w, http.StatusOK, jsonUser)
}

func (cfg *apiConfig) refresh(w http.ResponseWriter, r *http.Request) {
	type TokenType struct {
		Token string `json:"token"`
	}

	// Check for the token in the headers
	refreshToken, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Please provide an auth token", err)
		return
	}

	refreshTokenDB, err := cfg.db.GetRefreshToken(r.Context(), refreshToken)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "", err)
		return
	}

	// Check if the token has been revoked
	if refreshTokenDB.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "", err)
		return
	}

	// Start creating a new token for the authorized user
	expiration := time.Minute * 60

	token, err := auth.MakeJWT(refreshTokenDB.UserID, cfg.secret, expiration)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "JWT couldn't be generated", err)
		return
	}

	tokenResponse := &TokenType{
		Token: token,
	}

	respondWithJSON(w, http.StatusOK, tokenResponse)
}

func (cfg *apiConfig) revoke(w http.ResponseWriter, r *http.Request) {
	// Check for the token in the headers
	refreshToken, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Please provide an auth token", err)
		return
	}

	refreshTokenDB, err := cfg.db.GetRefreshToken(r.Context(), refreshToken)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "", err)
		return
	}

	sqlTimeNow := sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	refreshParams := &database.UpdateRefreshTokenParams{
		RevokedAt: sqlTimeNow,
		Token:     refreshTokenDB.Token,
	}

	_, err = cfg.db.UpdateRefreshToken(r.Context(), *refreshParams)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Refresh token couldn't be revoked", err)
		return
	}

	respondWithNoContent(w)
}

package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/tracevt/chirpy/internal/auth"
)

type UserData struct {
	UserID string `json:"user_id"`
}

type WebhookEvent struct {
	Event string   `json:"event"`
	Data  UserData `json:"data"`
}

func (cfg *apiConfig) parseEvent(w http.ResponseWriter, r *http.Request) {
	// Check for authorization
	providedKey, err := auth.GetAPIKey(r.Header)

	if err != nil || providedKey != cfg.polka {
		respondWithJSON(w, http.StatusUnauthorized, struct{}{})
	}

	decoder := json.NewDecoder(r.Body)
	event := WebhookEvent{}
	err = decoder.Decode(&event)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if event.Event != "user.upgraded" {
		respondWithNoContent(w)
		return
	}

	if event.Event == "user.upgraded" {
		// Upgrade the user
		userID, err := uuid.Parse(event.Data.UserID)

		if err != nil {
			respondWithNoContent(w)
			return
		}

		_, err = cfg.db.UpdateUserChirpyRed(r.Context(), userID)

		if err != nil {
			respondWithStatusCode(w, http.StatusNotFound)
		}

		respondWithJSON(w, http.StatusNoContent, struct{}{})
	}
}

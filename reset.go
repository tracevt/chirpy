package main

import (
	"net/http"
)

func (cfg *apiConfig) ResetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if !(cfg.platform == "" || cfg.platform == "dev") {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err := cfg.db.DropUsers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cfg.fileserverHits.Store(0)

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

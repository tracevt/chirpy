package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)

	hitsString := fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())

	io.WriteString(w, hitsString)
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}

func main() {
	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
	}
	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /healthz", HealthEndpoint)
	mux.HandleFunc("GET /metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /reset", apiCfg.ResetMetricsHandler)

	s := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Fatal(s.ListenAndServe())
}

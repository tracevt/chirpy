package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/tracevt/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Fatal("Error opening a database connection")
	}

	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		db:             database.New(db),
		platform:       platform,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", HealthEndpoint)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.ResetMetricsHandler)
	mux.HandleFunc("POST /api/validate_chirp", handleChirpsValidate)
	mux.HandleFunc("POST /api/users", apiCfg.createUser)

	s := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Fatal(s.ListenAndServe())
}

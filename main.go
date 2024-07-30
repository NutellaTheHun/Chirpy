package main

import (
	"internal/api"
	"internal/cDatabase"
	"log"
	"net/http"
)

func main() {

	dbPath := "database.json"
	cfg := &api.ApiConfig{}
	mux := http.NewServeMux()

	db, err := cDatabase.NewDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/app/*", cfg.MiddlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz", api.ReadyEndP)
	mux.HandleFunc("GET /admin/metrics", cfg.ServeAdminpage)
	mux.HandleFunc("/api/reset", cfg.ResetHits)

	mux.HandleFunc("GET /api/chirps", db.HandleGetChirpsRequest)
	mux.HandleFunc("POST /api/chirps", db.HandlePostChirpsRequest)

	ser := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	ser.ListenAndServe()

}

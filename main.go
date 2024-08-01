package main

import (
	"internal/api"
	"internal/cDatabase"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load("config.env")
	if err != nil {
		log.Print("godotenv error: ", err.Error())
	}
	jwtSecret := os.Getenv("JWT_SECRET")

	dbPath := "database.json"
	db, err := cDatabase.NewDB(dbPath, jwtSecret)
	if err != nil {
		log.Fatal(err)
	}

	cfg := &api.ApiConfig{}

	mux := http.NewServeMux()

	mux.Handle("/app/*", cfg.MiddlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz", api.ReadyEndP)

	mux.HandleFunc("GET /admin/metrics", cfg.ServeAdminpage)
	mux.HandleFunc("/api/reset", cfg.ResetHits)

	mux.HandleFunc("GET /api/chirps", db.HandleGetChirpsRequest)
	mux.HandleFunc("GET /api/chirps/{chirpId}", db.HandleGetChirpRequest)
	mux.HandleFunc("POST /api/chirps", db.HandlePostChirpsRequest)

	mux.HandleFunc("POST /api/users", db.HandlePostUsers)
	mux.HandleFunc("PUT /api/users", db.HandlePutUsersRequest)

	mux.HandleFunc("POST /api/login", db.HandleLoginRequest)

	mux.HandleFunc("POST /api/refresh", db.HandlePostRefresh)

	mux.HandleFunc("POST /api/revoke", db.HandlePostRevoke)

	ser := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	ser.ListenAndServe()

}

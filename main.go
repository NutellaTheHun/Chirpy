package main

import (
	"net/http"
)

func main() {

	cfg := &apiConfig{}
	mux := http.NewServeMux()
	mux.Handle("/app/*", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz", readyEndP)
	//mux.HandleFunc("GET /admin/metrics", cfg.requestNum)
	mux.HandleFunc("GET /admin/metrics", cfg.serveAdminpage)
	mux.HandleFunc("/api/reset", cfg.resetHits)
	//mux.HandleFunc("POST /api/validate_chirp", validateChirp)
	ser := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	ser.ListenAndServe()
}

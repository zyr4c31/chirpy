package main

import (
	"net/http"
)

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Cache-Control", "no-cache")
		if r.Method == "Options" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type apiConfig struct {
	fileserverHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type:", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) metrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type:", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits++
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type:", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits = 0
	w.Write([]byte("OK"))
}

func main() {
	port := (":8080")

	api := apiConfig{
		fileserverHits: 0,
	}

	mux := http.NewServeMux()

	corsMux := api.middlewareMetricsInc(mux)

	s := &http.Server{
		Addr:    port,
		Handler: corsMux,
	}

	fsh := http.FileServer(http.Dir("."))
	sph := http.StripPrefix("/app/", api.middlewareMetricsInc(fsh))

	mux.Handle("/app/", api.middlewareMetricsInc(sph))

	mux.HandleFunc("/healthz", healthz)

	mux.HandleFunc("/metrics", api.metrics)
	mux.HandleFunc("/reset", api.reset)

	s.ListenAndServe()
}

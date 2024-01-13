package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Cache-control", "no-cache")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r != nil {
			w.WriteHeader(http.StatusOK)
		}
		next.ServeHTTP(w, r)
	})
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

type apiConfig struct {
	fileserverHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
}

func (cfg *apiConfig) metrics(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hits: " + fmt.Sprint(cfg.fileserverHits)))
}

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	port := "8080"

	// filerootPath := "."

	apicfg := &apiConfig{
		fileserverHits: 0,
	}

	r := chi.NewRouter()

	mux := http.NewServeMux()

	fsHandler := http.FileServer(http.Dir("."))
	spHandler := http.StripPrefix("/app", fsHandler)

	r.Handle("/app/", apicfg.middlewareMetricsInc(spHandler))
	r.Get("/healthz", healthz)
	mux.HandleFunc("/reset", apicfg.reset)
	mux.HandleFunc("/metrics", apicfg.metrics)

	corsMux := middlewareCors(r)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	srv.ListenAndServe()
}

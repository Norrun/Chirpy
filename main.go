package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"
)

func main() {
	log.Println("setting up")
	var conf apiConfig
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app",
		conf.middlewareMetricsInc(
			http.FileServer(http.Dir(".")))))
	mux.HandleFunc("/healthz", readinessHandler)

	mux.HandleFunc("/metrics", conf.handlerHitCount)

	mux.HandleFunc("/reset", conf.handlerHitCountReset)

	server := http.Server{Handler: mux,
		Addr: ":8080"}

	log.Println("serving")
	log.Fatal(server.ListenAndServe())

}

func readinessHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	io.WriteString(res, "OK")

}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerHitCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, fmt.Sprintf("Hits: %x", cfg.fileserverHits.Load()))
}

func (cfg *apiConfig) handlerHitCountReset(res http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	io.WriteString(res, "OK")
}

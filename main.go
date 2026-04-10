package main

import (
	"io"
	"log"
	"net/http"
)

func main() {
	log.Println("setting up")
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/healthz", readinessHandler)
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

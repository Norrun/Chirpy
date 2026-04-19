package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil && code >= 500 {
		log.Printf("5XX error (%s): %v", msg, err)
	}
	log.Print(err)

	res := struct {
		Error string `json:"error"`
	}{Error: msg}
	respondWithJson(w, code, res)
}

func respondWithJson(w http.ResponseWriter, code int, body any) {
	resb, err := json.Marshal(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Unable to generate json reponse: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(resb)
}

func readJsonRequestType[T any](r *http.Request) (T, error) {
	var obj T
	err := readJsonRequest(r, &obj)
	if err != nil {
		return obj, err
	}
	return obj, nil
}

func readJsonRequest(r *http.Request, to any) error {
	bodb, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bodb, to)
	if err != nil {
		return err
	}
	return nil
}

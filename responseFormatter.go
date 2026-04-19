package main

import (
	"log"
	"net/http"

	"github.com/Norrun/Chirpy/internal/errormeta"
	"github.com/Norrun/Chirpy/internal/renderstuff"
)

func middlewareRespondJson[T any](conf *apiConfig, h renderstuff.Handler[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := h(r)
		if err != nil {
			status, _ := errormeta.Has[int](err)
			msg, _ := errormeta.Has[string](err)

			if status == 0 {
				status = 500
			}
			if msg == "" {
				msg = "something went wrong"
			}

			if conf.Platform == "dev" {
				log.Println(err)
			} else if status >= 500 {
				log.Printf("5XX error (%s): %v", msg, err)
			}

			res := struct {
				Error string `json:"error"`
			}{Error: msg}
			respondWithJson(w, status, res)
			return
		}
		for key, v := range res.Headers {
			for _, e := range v {
				w.Header().Add(key, e)
			}

		}
		status := res.Status
		if status == 0 {
			status = 200
		}
		respondWithJson(w, status, res.Data)

	}
}

package main

import (
	"log"
	"net/http"

	"github.com/Norrun/Chirpy/internal/errormeta"
	"github.com/Norrun/Chirpy/internal/flexy"
)

//type serializable any

func middlewareRespondJson[T any](conf *apiConfig, h flexy.HandlerFunc[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := h(r)
		if err != nil {
			status, _ := errormeta.Extract[int](err)
			msg, _ := errormeta.Extract[string](err)

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
		if res.Headers != nil {
			for key, v := range res.Headers {
				for _, e := range v {
					w.Header().Add(key, e)
				}

			}
		}

		status := res.Status
		if status == 0 {
			status = 200
		}
		respondWithJson(w, status, res.Data)

	}
}

func adapterHandleError(conf *apiConfig, previous flexy.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := previous.ServeHTTP(w, r)
		if err != nil {
			status, _ := errormeta.Extract[int](err)
			msg, _ := errormeta.Extract[string](err)

			if status == 0 {
				status = 500
			}
			if msg == "" {
				msg = "something went wrong"
			}

			if conf.Platform == "dev" {
				log.Printf("%d_%s_%v", status, msg, err)
			} else if status >= 500 {
				log.Printf("5XX error (%s): %v", msg, err)
			}

			res := struct {
				Error string `json:"error"`
			}{Error: msg}
			respondWithJson(w, status, res)
			return
		}
	})

}

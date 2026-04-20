package main

import (
	"io"
	"log"
	"net/http"
	"strings"

	_ "github.com/lib/pq"
)

//type httpStatus int
//type userMessage string

func main() {
	log.Println("setting up")
	conf, err := loading()
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	routing(mux, conf)

	server := http.Server{Handler: mux,
		Addr: ":8080"}

	log.Println("serving")
	log.Fatal(server.ListenAndServe())

}

// Routing
func routing(mux *http.ServeMux, conf *apiConfig) {

	mux.Handle("/app/", http.StripPrefix("/app",
		conf.middlewareMetricsInc(
			http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", handlerGETapiHealth)
	mux.HandleFunc("GET /admin/metrics", conf.handlerAdminMetrics)
	mux.HandleFunc("POST /admin/reset", conf.handlerAdminReset)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)
	mux.HandleFunc("POST /api/users", conf.handlerApiUsersCreate)
	mux.HandleFunc("POST /api/chirps", conf.handlerApiChirpsCreate)
	mux.Handle("GET /api/chirps", middlewareRespondJson(conf, conf.handlerApiChirps))

}

// Handlers (and stuf)

func handlerGETapiHealth(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	io.WriteString(res, "OK")

}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	post, err := readJsonRequestType[Request](r)

	if err != nil {
		log.Println(err)
		respondWithError(w, 400, "something went wrong", err)
	}

	if len([]rune(post.Body)) > 140 {
		respondWithError(w, 400, "Chirp is too long", nil)
		return
	}
	valid := ResponseClean{CleanedBody: stringCleaner(post.Body)}
	respondWithJson(w, 200, valid)

}

func stringCleaner(s string) string {
	words := strings.Split(s, " ")
	cleanWords := make([]string, 0, len(words))

	for _, w := range words {
		switch strings.ToLower(w) {
		case "kerfuffle", "sharbert", "fornax":
			cleanWords = append(cleanWords, "****")
		default:
			cleanWords = append(cleanWords, w)
		}
	}
	return strings.Join(cleanWords, " ")
}

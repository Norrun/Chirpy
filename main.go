package main

import (
	"database/sql"
	"encoding/json"
	"os"

	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/Norrun/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

//type httpStatus int
//type userMessage string

func main() {
	log.Println("setting up")

	log.Fatal(godotenv.Load())
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	queries := database.New(db)

	conf := apiConfig{dbq: queries}
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app",
		conf.middlewareMetricsInc(
			http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", readinessHandler)

	mux.HandleFunc("GET /admin/metrics", conf.handlerHitCount)

	mux.HandleFunc("POST /admin/reset", conf.handlerHitCountReset)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

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
	dbq            *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerHitCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, fmt.Sprintf("<html>\n  <body>\n    <h1>Welcome, Chirpy Admin</h1>\n    <p>Chirpy has been visited %d times!</p>\n  </body>\n</html>", cfg.fileserverHits.Load()))
}

func (cfg *apiConfig) handlerHitCountReset(res http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	io.WriteString(res, "OK")
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	var post Request
	bodb, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		respondWithError(w, 400, "something went wrong")
		return
	}
	err = json.Unmarshal(bodb, &post)
	if err != nil {
		log.Println(err)
		respondWithError(w, 400, "something went wrong")
		return
	}

	if len([]rune(post.Body)) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	valid := ResponseClean{CleanedBody: stringCleaner(post.Body)}
	respondWithJson(w, 200, valid)

}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	res := ResponseErr{Error: msg}
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

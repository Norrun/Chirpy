package main

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Norrun/Chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func loading() (*apiConfig, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}
	queries := database.New(db)

	conf := apiConfig{dbq: queries, Platform: platform}
	return &conf, nil
}

type apiConfig struct {
	fileserverHits atomic.Int32
	dbq            *database.Queries
	Platform       string
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

func (cfg *apiConfig) handlerReset(res http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)

	if cfg.Platform != "dev" {
		respondWithError(res, http.StatusForbidden, "forbidden", nil)

		return
	}
	err := cfg.dbq.Reset(req.Context())
	if err != nil {
		respondWithError(res, 500, "reset failed", err)

		return
	}

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	io.WriteString(res, "OK")
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	var createUser struct {
		Email string `json:"email"`
	}

	err := readJsonRequest(r, &createUser)
	if err != nil {
		respondWithError(w, 400, "failed to create user", err)
		return
	}

	user, err := cfg.dbq.CreateUser(r.Context(), createUser.Email)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			respondWithError(w, 400, "User already exists", err)
		} else {
			respondWithError(w, 500, "something went wrong when creating user", err)
		}
		return
	}
	responseUser := struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	respondWithJson(w, http.StatusCreated, responseUser)

}

package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Norrun/Chirpy/internal/auth"
	"github.com/Norrun/Chirpy/internal/database"
	"github.com/Norrun/Chirpy/internal/errormeta"
	"github.com/Norrun/Chirpy/internal/flexy"
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
	secret := os.Getenv("SECRET")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}
	queries := database.New(db)

	conf := apiConfig{dbq: queries, Platform: platform, Secret: secret}
	return &conf, nil
}

type apiConfig struct {
	fileserverHits atomic.Int32
	dbq            *database.Queries
	Platform       string
	Secret         string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerAdminMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, fmt.Sprintf("<html>\n  <body>\n    <h1>Welcome, Chirpy Admin</h1>\n    <p>Chirpy has been visited %d times!</p>\n  </body>\n</html>", cfg.fileserverHits.Load()))
}

func (cfg *apiConfig) handlerAdminReset(res http.ResponseWriter, req *http.Request) {
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

func (cfg *apiConfig) handlerApiUsersCreate(w http.ResponseWriter, r *http.Request) {

	createUser, err := readJsonRequestType[CreateUser](r)
	if err != nil {
		respondWithError(w, 400, "failed to create user", err)
		return
	}

	password, err := auth.HashPassword(createUser.Password)
	if err != nil {
		respondWithError(w, 500, "internal error", err)
		return
	}

	user, err := cfg.dbq.CreateUser(r.Context(), database.CreateUserParams{Email: createUser.Email, Password: password})

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			respondWithError(w, 400, "User already exists", err)
		} else {
			respondWithError(w, 500, "something went wrong when creating user", err)
		}
		return
	}
	responseUser := ResponseUser{
		ID:        user.ID.String(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	respondWithJson(w, http.StatusCreated, responseUser)

}

func (conf *apiConfig) handlerApiChirpsCreate(w http.ResponseWriter, r *http.Request) {

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Unauthorized", err)
		return
	}
	userID, err := auth.ValidateJWT(token, conf.Secret)
	if err != nil {
		respondWithError(w, 401, "Unauthorized", err)
		return
	}

	var recevePost struct {
		Body string `json:"body"`
	}
	err = readJsonRequest(r, &recevePost)

	if err != nil {
		respondWithError(w, 400, "Something went wrong when creating post", err)
		return
	}
	if len([]rune(recevePost.Body)) > 140 {
		respondWithError(w, 400, "Chirp is too long", nil)
		return
	}

	post, err := conf.dbq.CreatePost(r.Context(), database.CreatePostParams{Body: recevePost.Body, UserID: userID})
	if err != nil {
		respondWithError(w, 400, "Something went wrong when creating post", err)
		return
	}

	responsePost := ResponseChirp{
		ID:        post.ID.String(),
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
		Body:      stringCleaner(post.Body),
		UserID:    post.UserID.String(),
	}

	respondWithJson(w, http.StatusCreated, responsePost)

}

func (receiver *apiConfig) handlerApiChirps(r *http.Request) (flexy.HandlerResult[[]ResponseChirp], error) {
	var empty flexy.HandlerResult[[]ResponseChirp]
	chirps, err := receiver.dbq.GetPosts(r.Context())
	if err != nil {
		return empty, err
	}
	resChirps := make([]ResponseChirp, 0, len(chirps))
	for _, chirp := range chirps {
		resChirps = append(resChirps, ResponseChirp{
			ID:        chirp.ID.String(),
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.ID.String(),
		})
	}
	return flexy.HandlerResult[[]ResponseChirp]{Status: 200, Headers: nil, Data: resChirps}, nil
}

func (receiver *apiConfig) handlerApiChirpsID(r *http.Request) (flexy.HandlerResult[ResponseChirp], error) {
	var empty flexy.HandlerResult[ResponseChirp]
	IDStr := r.PathValue("id")
	ID, err := uuid.Parse(IDStr)
	if err != nil {
		err = errormeta.Include(err, 404)
		err = errormeta.Include(err, "Not Found")
		return empty, err
	}
	chirp, err := receiver.dbq.GetPost(r.Context(), ID)
	if err != nil {
		if err == sql.ErrNoRows {
			err = errormeta.Include(err, 404)
			err = errormeta.Include(err, "Chirp Not Found")
			return empty, err
		}
		return empty, err
	}

	resChirp := ResponseChirp{
		ID:        IDStr,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID.String(),
	}

	return flexy.HandlerResult[ResponseChirp]{Data: resChirp}, nil
}

func (receiver *apiConfig) handlerApiLogin(r *http.Request) (flexy.HandlerResult[ResponseUser], error) {
	var empty flexy.HandlerResult[ResponseUser]
	login, err := readJsonRequestType[CreateUser](r)
	if err != nil {
		if _, ok := errors.AsType[*json.UnmarshalTypeError](err); ok {
			err = errormeta.Include(err, 422)
			err = errormeta.Include(err, "need email and password fields")
		} else {
			err = errormeta.Include(err, 400)
			err = errormeta.Include(err, "login error")
		}
		return empty, err
	}
	user, err := receiver.dbq.GetUserByEmail(r.Context(), login.Email)
	if err != nil {
		err = errormeta.Include(err, 400)
		err = errormeta.Include(err, "login error")
		return empty, err
	}
	valid, err := auth.CheckPasswordHash(login.Password, user.Password)
	if err != nil {
		err = errormeta.Include(err, "internal server error")
		return empty, err
	}
	if valid == false {
		err = fmt.Errorf("issue")
		err = errormeta.Include(err, 401)
		err = errormeta.Include(err, "wrong password")
		return empty, err
	}
	//TODO: remove hardcoded value
	expires := time.Hour

	jwt, err := auth.MakeJWT(user.ID, receiver.Secret, expires)
	if err != nil {
		return empty, err
	}
	refresh := auth.MakeRefreshToken()

	receiver.dbq.RegisterRefreshToken(r.Context(), database.RegisterRefreshTokenParams{
		Token:     refresh,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	})

	result := flexy.HandlerResult[ResponseUser]{Data: ResponseUser{
		ID:           user.ID.String(),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        jwt,
		RefreshToken: refresh,
	}}
	return result, nil
}

func (receiver *apiConfig) handlerApiRefresh(w http.ResponseWriter, r *http.Request) error {
	barer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		err = errormeta.Include(err, 400)
		err = errormeta.Include(err, "Bad Request")
		return err
	}
	refresh, err := receiver.dbq.GetRefreshToken(r.Context(), barer)
	if err != nil {
		if err == sql.ErrNoRows {
			err = errormeta.Include(err, 401)
			err = errormeta.Include(err, "Unauthorized")
		}
		return err
	}
	if refresh.ExpiresAt.Before(time.Now()) || refresh.RevokedAt.Valid {
		err = errormeta.Include(err, 401)
		err = errormeta.Include(err, "Unauthorized")
		return err
	}
	token, err := auth.MakeJWT(refresh.UserID, receiver.Secret, time.Hour)
	if err != nil {
		return err
	}
	tokenjson := struct {
		Token string ` json:"token"`
	}{
		token,
	}
	respondWithJson(w, 200, tokenjson)
	return nil
}

func (receiver *apiConfig) handlerApiRevoke(w http.ResponseWriter, r *http.Request) error {
	barer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		err = errormeta.Include(err, 400)
		err = errormeta.Include(err, "Bad Request")
		return err
	}
	err = receiver.dbq.RevokeRefreshToken(r.Context(), barer)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (receiver *apiConfig) handlerApiUsersUpdate(w http.ResponseWriter, r *http.Request) error {
	barer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		err = errormeta.Include(err, 401)
		err = errormeta.Include(err, "Unauthorized")
		return err
	}
	id, err := auth.ValidateJWT(barer, receiver.Secret)

	if err != nil {
		err = errormeta.Include(err, 401)
		err = errormeta.Include(err, "Unauthorized")
		return err
	}

	update, err := readJsonRequestType[CreateUser](r)
	if err != nil {
		err = errormeta.Include(err, 400)
		err = errormeta.Include(err, "Bad Request")
		return err
	}
	passhash, err := auth.HashPassword(update.Password)
	if err != nil {

		return err
	}
	user, err := receiver.dbq.ChangeCredentials(r.Context(), database.ChangeCredentialsParams{
		ID:       id,
		Email:    update.Email,
		Password: passhash,
	})

	if err != nil {
		err = errormeta.Include(err, 401)
		err = errormeta.Include(err, "Unauthorized")
		return err
	}
	response := ResponseUser{
		ID:        id.String(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	respondWithJson(w, 200, response)
	return nil
}

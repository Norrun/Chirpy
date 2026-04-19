package main

import "net/http"

type HandlerResult[T any] struct {
	Status  int
	Headers http.Header
	Data    T
}
type ReadRequest func(*http.Request, any) error
type ReaderRequestType[T any] func(*http.Request) (T, error)
type Handler[T any] func(*http.Request) (HandlerResult[T], error)

func RenderResponce[T any](conf *apiConfig, h Handler[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := h(r)
	}
}

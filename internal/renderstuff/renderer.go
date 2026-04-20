package renderstuff

import "net/http"

type HandlerResult[T any] struct {
	Status  int
	Headers http.Header
	Data    T
}
type RequestReader func(*http.Request, any) error

func ReadRequestType[T any](reader RequestReader, r *http.Request) (T, error) {
	var res T
	err := reader(r, &res)
	return res, err
}

type Handler[T any] func(*http.Request, ...RequestReader) (HandlerResult[T], error)

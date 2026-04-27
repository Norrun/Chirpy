package flexy

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

type HandlerFunc[T any] func(*http.Request) (HandlerResult[T], error)

type Renderer interface {
	Render(http.ResponseWriter, int, any, *http.Request) error
}

type RendererFunc func(http.ResponseWriter, int, any, *http.Request) error

func (receiver RendererFunc) Render(w http.ResponseWriter, status int, a any, r *http.Request) error {
	return receiver(w, status, a, r)
}

type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request) error
}

type Handlerf func(w http.ResponseWriter, r *http.Request) error

func (receiver Handlerf) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return receiver(w, r)
}

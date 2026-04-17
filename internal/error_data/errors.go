package errordata

import (
	"errors"
	"fmt"
)

type ErrorMetadata[T any] interface {
	error
	Metadata() T
	Unwrap() error
}

type DefaultErrorMetadata[T any] struct {
	internal error
	data     T
}

func (r DefaultErrorMetadata[T]) Error() string {
	if v, ok := any(r.data).(fmt.Stringer); ok {
		return fmt.Sprintf("%s\n%s", r.internal.Error(), v.String())
	}

	return r.internal.Error()
}

func (r DefaultErrorMetadata[T]) Metadata() T {
	return r.data
}

func (r DefaultErrorMetadata[T]) Unwrap() error {
	return r.internal
}

func WithMetadata[T any](err error, data T) DefaultErrorMetadata[T] {
	return DefaultErrorMetadata[T]{internal: err, data: data}
}

func WithMetadataUnique[T any](err error, data T) error {
	if _, ok := errors.AsType[ErrorMetadata[T]](err); ok {
		return err
	}
	return DefaultErrorMetadata[T]{internal: err, data: data}
}

func Having[T any](err error) (T, bool) {

	data, ok := errors.AsType[ErrorMetadata[T]](err)
	if ok {
		return data.Metadata(), true
	}
	var temp T
	return temp, false
}

func Test() {
	var err error
	Having[[2]int](err)
}

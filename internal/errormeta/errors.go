package errormeta

import (
	"errors"
	"fmt"
)

type ErrorMetadata[T any] interface {
	error
	Metadata() T
	Unwrap() error
}

type errorMetadata[T any] struct {
	internal error
	data     T
}

func (r errorMetadata[T]) Error() string {

	if r.internal == nil {
		return fmt.Sprintf("%v", r.data)
	}

	if v, ok := any(r.data).(fmt.Stringer); ok {
		return fmt.Sprintf("%s\n%s", r.internal.Error(), v.String())
	}
	if v, ok := any(r.data).(string); ok {
		return fmt.Sprintf("%s\n%s", r.internal.Error(), v)
	}

	return r.internal.Error()
}

func (r errorMetadata[T]) Metadata() T {
	return r.data
}

func (r errorMetadata[T]) Unwrap() error {
	return r.internal
}

func Include[T any](err error, data T) ErrorMetadata[T] {
	return errorMetadata[T]{internal: err, data: data}
}

func IncludeIfFirstOfType[T any](err error, data T) error {
	if _, ok := errors.AsType[ErrorMetadata[T]](err); ok {
		return err
	}
	return errorMetadata[T]{internal: err, data: data}
}

func Has[T any](err error) bool {
	_, ok := errors.AsType[ErrorMetadata[T]](err)
	return ok
}

func Extract[T any](err error) (T, bool) {

	data, ok := errors.AsType[ErrorMetadata[T]](err)
	if ok {
		return data.Metadata(), true
	}
	var temp T
	return temp, false
}

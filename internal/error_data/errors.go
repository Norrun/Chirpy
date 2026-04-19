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

type errorMetadata[T any] struct {
	internal error
	data     T
}

func (r errorMetadata[T]) Error() string {
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

func WithMetadata[T any](err error, data T) ErrorMetadata[T] {
	return errorMetadata[T]{internal: err, data: data}
}

func WithMetadataUnique[T any](err error, data T) error {
	if _, ok := errors.AsType[ErrorMetadata[T]](err); ok {
		return err
	}
	return errorMetadata[T]{internal: err, data: data}
}

func Extract[T any](err error) (T, bool) {

	data, ok := errors.AsType[ErrorMetadata[T]](err)
	if ok {
		return data.Metadata(), true
	}
	var temp T
	return temp, false
}

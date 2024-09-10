package usecase

import "errors"

var (
	ErrNotAMethod     = errors.New("selected element is not a method")
	ErrNotImplemented = errors.New("not implemented")
)

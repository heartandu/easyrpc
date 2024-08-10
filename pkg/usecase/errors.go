package usecase

import "errors"

var (
	ErrDescriptorNotFound = errors.New("descriptor not found")
	ErrNotAMethod         = errors.New("selected element is not a method")
	ErrNotImplemented     = errors.New("not implemented")
	ErrInvalidFQN         = errors.New("invalid fully qualified method name")
)

package cmds

import "errors"

var (
	ErrMissingArgs    = errors.New("missing arguments")
	ErrUnexpectedArgs = errors.New("unexpected arguments")
	ErrValidation     = errors.New("validation failed")
)

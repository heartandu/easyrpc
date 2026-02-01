package cmds

import "errors"

var (
	ErrMissingArgs = errors.New("missing arguments")
	ErrValidation  = errors.New("validation failed")
)

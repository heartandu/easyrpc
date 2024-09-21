package cmds

import "errors"

var (
	ErrMissingArgs      = errors.New("missing arguments")
	ErrValidation       = errors.New("validation failed")
	ErrMissingCertOrKey = errors.New("cert and key must be both set")
	ErrEmptyAddress     = errors.New("address must not be empty")
	ErrNoSource         = errors.New("at least 1 proto file must be specified or reflection used")
)

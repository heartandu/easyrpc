package config

import "errors"

var (
	ErrMissingCertOrKey = errors.New("cert and key must be both set")
	ErrEmptyAddress     = errors.New("address must not be empty")
	ErrNoSource         = errors.New("at least 1 proto file must be specified, imported all files or reflection used")
)

// Validator encapsulates the logic of validating application configuration for specific scenarios.
type Validator struct {
	validateConn bool
}

// NewValidator returns a new instance of ConfigValidator.
func NewValidator(opts ...ValidatorOption) *Validator {
	v := &Validator{}

	for _, o := range opts {
		o(v)
	}

	return v
}

// ValidatorOption modifies the behavior of ConfigValidator.
type ValidatorOption func(v *Validator)

// WithValidateConn makes validator to check connection options.
func WithValidateConn(val bool) ValidatorOption {
	return func(v *Validator) {
		v.validateConn = val
	}
}

// Validate checks that call configuration was set up correctly.
func (v *Validator) Validate(c *Config) (err error) {
	if len(c.Proto.ProtoFiles) == 0 && !c.Proto.ImportAll && !c.Server.Reflection {
		err = errors.Join(err, ErrNoSource)
	}

	if v.validateConn {
		if c.Server.Address == "" {
			err = errors.Join(err, ErrEmptyAddress)
		}

		if c.TLS.Cert == "" && c.TLS.Key != "" || c.TLS.Cert != "" && c.TLS.Key == "" {
			err = errors.Join(err, ErrMissingCertOrKey)
		}
	}

	return err
}

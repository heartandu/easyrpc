package fqn

import "strings"

const partsDelim = "."

// FullyQualifiedMethodName returns a fully qualified method name.
// If method is missing package or service names, then defaultPackage and defaultService will be used instead.
// If method is empty, empty string is returned.
func FullyQualifiedMethodName(method, defaultPackage, defaultService string) string {
	if method == "" {
		return ""
	}

	fullyQualifiedMethodName := ParseFQMN(method)

	if fullyQualifiedMethodName.PackageName == "" {
		fullyQualifiedMethodName.PackageName = defaultPackage
	}

	if fullyQualifiedMethodName.Service == "" {
		fullyQualifiedMethodName.Service = defaultService
	}

	return fullyQualifiedMethodName.PartsBuilder().
		WithPackage().
		WithService().
		String()
}

// ParseFQMN parses and returns parts of a fully qualified method name.
func ParseFQMN(method string) FQMN {
	parts := strings.Split(method, partsDelim)
	partsLen := len(parts)

	switch partsLen {
	case 1:
		return FQMN{
			Method: parts[0],
		}
	case 2:
		return FQMN{
			Service: parts[0],
			Method:  parts[1],
		}
	default:
		return FQMN{
			PackageName: strings.Join(parts[:partsLen-2], partsDelim),
			Service:     parts[partsLen-2],
			Method:      parts[partsLen-1],
		}
	}
}

// FQMN is a representation of a fully qualified method name split into parts.
type FQMN struct {
	PackageName string
	Service     string
	Method      string
}

// PartsBuilder returns a builder object which controls which parts of fully qualified method name should be rendered.
func (mn FQMN) PartsBuilder() *FQMNBuilder {
	return &FQMNBuilder{fqmn: &mn}
}

// FQMNBuilder builds fully qualified method name strings.
// To maintain correctness:
//   - if WithPackage is called and the service name is empty, only the method name is rendered
//   - calling WithPackage also causes the service name to be rendered (even without WithService)
type FQMNBuilder struct {
	fqmn        *FQMN
	withPackage bool
	withService bool
}

// WithPackage marks the package name to be rendered.
// Forces to render the service name to maintain fully qualified method name correctness.
func (b *FQMNBuilder) WithPackage() *FQMNBuilder {
	b.withPackage = true

	return b
}

// WithService marks the service name to be rendered.
func (b *FQMNBuilder) WithService() *FQMNBuilder {
	b.withService = true

	return b
}

func (b *FQMNBuilder) String() string {
	var sb strings.Builder

	if b.withPackage && b.fqmn.PackageName != "" && b.fqmn.Service != "" {
		sb.WriteString(b.fqmn.PackageName)
	}

	if (b.withPackage || b.withService) && b.fqmn.Service != "" {
		if sb.Len() > 0 {
			sb.WriteString(partsDelim)
		}

		sb.WriteString(b.fqmn.Service)
	}

	if sb.Len() > 0 {
		sb.WriteString(partsDelim)
	}

	sb.WriteString(b.fqmn.Method)

	return sb.String()
}

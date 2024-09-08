package fqn

import (
	"fmt"
	"strings"
)

// FullyQualifiedMethodName returns a fully qualified method name.
// If method is missing package or service names, then defaultPackage and defaultService will be used instead.
// If method is empty, empty string is returned.
func FullyQualifiedMethodName(method, defaultPackage, defaultService string) string {
	if method == "" {
		return ""
	}

	fullyQualifiedMethodName := parseFQMN(method)

	if fullyQualifiedMethodName.packageName == "" {
		fullyQualifiedMethodName.packageName = defaultPackage
	}

	if fullyQualifiedMethodName.service == "" {
		fullyQualifiedMethodName.service = defaultService
	}

	return fullyQualifiedMethodName.String()
}

func parseFQMN(method string) fqmn {
	const minFQMNPartsLen = 3

	parts := strings.Split(method, ".")
	partsLen := len(parts)

	if partsLen == 1 {
		return fqmn{
			method: parts[0],
		}
	}

	if partsLen < minFQMNPartsLen {
		return fqmn{
			service: getOrDefault(parts, 0, ""),
			method:  getOrDefault(parts, 1, ""),
		}
	}

	return fqmn{
		packageName: strings.Join(parts[:partsLen-2], "."),
		service:     parts[partsLen-2],
		method:      parts[partsLen-1],
	}
}

func getOrDefault(parts []string, index int, def string) string {
	if len(parts) < index+1 {
		return def
	}

	return parts[index]
}

type fqmn struct {
	packageName string
	service     string
	method      string
}

func (mn *fqmn) String() string {
	return fmt.Sprintf("%s.%s.%s", mn.packageName, mn.service, mn.method)
}

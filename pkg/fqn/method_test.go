package fqn_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/heartandu/easyrpc/pkg/fqn"
)

func TestFullyQualifiedMethodName(t *testing.T) {
	t.Parallel()

	type args struct {
		method         string
		defaultPackage string
		defaultService string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "method is fully qualified",
			args: args{
				method:         "service.v1.Service.Method",
				defaultPackage: "test.v1",
				defaultService: "Test",
			},
			want: "service.v1.Service.Method",
		},
		{
			name: "method is a service.method",
			args: args{
				method:         "Service.Method",
				defaultPackage: "test.v1",
				defaultService: "Test",
			},
			want: "test.v1.Service.Method",
		},
		{
			name: "method is a method",
			args: args{
				method:         "Method",
				defaultPackage: "test.v1",
				defaultService: "Test",
			},
			want: "test.v1.Test.Method",
		},
		{
			name: "empty method",
			args: args{
				method:         "",
				defaultPackage: "test.v1",
				defaultService: "Test",
			},
			want: "",
		},
		{
			name: "fully qualified method without defaults",
			args: args{
				method: "service.v1.Service.Method",
			},
			want: "service.v1.Service.Method",
		},
		{
			name: "service.method without defaults",
			args: args{
				method: "Service.Method",
			},
			want: "Service.Method",
		},
		{
			name: "method without defaults",
			args: args{
				method: "Method",
			},
			want: "Method",
		},
		{
			name: "method with default service",
			args: args{
				method:         "Method",
				defaultService: "Service",
			},
			want: "Service.Method",
		},
		{
			name: "method with default package",
			args: args{
				method:         "Method",
				defaultPackage: "test.v1",
			},
			want: "Method",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := fqn.FullyQualifiedMethodName(tt.args.method, tt.args.defaultPackage, tt.args.defaultService)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParseFQMN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		want   fqn.FQMN
	}{
		{
			name:   "empty string",
			method: "",
			want:   fqn.FQMN{},
		},
		{
			name:   "method only",
			method: "Method",
			want: fqn.FQMN{
				Method: "Method",
			},
		},
		{
			name:   "service and method",
			method: "Service.Method",
			want: fqn.FQMN{
				Service: "Service",
				Method:  "Method",
			},
		},
		{
			name:   "package service and method",
			method: "test.v1.Service.Method",
			want: fqn.FQMN{
				PackageName: "test.v1",
				Service:     "Service",
				Method:      "Method",
			},
		},
		{
			name:   "deeply nested package",
			method: "a.b.c.d.Service.Method",
			want: fqn.FQMN{
				PackageName: "a.b.c.d",
				Service:     "Service",
				Method:      "Method",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := fqn.ParseFQMN(tt.method)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestFQMNBuilder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		method      string
		withPackage bool
		withService bool
		want        string
	}{
		{
			name:   "empty string, no flags",
			method: "",
			want:   "",
		},
		{
			name:        "empty string, with package",
			method:      "",
			withPackage: true,
			want:        "",
		},
		{
			name:        "empty string, with service",
			method:      "",
			withService: true,
			want:        "",
		},
		{
			name:   "method only, no flags",
			method: "Method",
			want:   "Method",
		},
		{
			name:        "method only, with service",
			method:      "Method",
			withService: true,
			want:        "Method",
		},
		{
			name:        "method only, with package",
			method:      "Method",
			withPackage: true,
			want:        "Method",
		},
		{
			name:        "method only, with both flags",
			method:      "Method",
			withPackage: true,
			withService: true,
			want:        "Method",
		},
		{
			name:   "service and method, no flags",
			method: "Service.Method",
			want:   "Method",
		},
		{
			name:        "service and method, with service",
			method:      "Service.Method",
			withService: true,
			want:        "Service.Method",
		},
		{
			name:        "service and method, with package",
			method:      "Service.Method",
			withPackage: true,
			want:        "Service.Method",
		},
		{
			name:        "service and method, with both flags",
			method:      "Service.Method",
			withPackage: true,
			withService: true,
			want:        "Service.Method",
		},
		{
			name:   "fully qualified, no flags",
			method: "test.v1.Service.Method",
			want:   "Method",
		},
		{
			name:        "fully qualified, with service",
			method:      "test.v1.Service.Method",
			withService: true,
			want:        "Service.Method",
		},
		{
			name:        "fully qualified, with package",
			method:      "test.v1.Service.Method",
			withPackage: true,
			want:        "test.v1.Service.Method",
		},
		{
			name:        "fully qualified, with both flags",
			method:      "test.v1.Service.Method",
			withPackage: true,
			withService: true,
			want:        "test.v1.Service.Method",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b := fqn.ParseFQMN(tt.method).PartsBuilder()
			if tt.withPackage {
				b.WithPackage()
			}

			if tt.withService {
				b.WithService()
			}

			require.Equal(t, tt.want, b.String())
		})
	}
}

package fqn_test

import (
	"testing"

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := fqn.FullyQualifiedMethodName(tt.args.method, tt.args.defaultPackage, tt.args.defaultService)
			if got != tt.want {
				t.Fatalf("FullyQualifiedMethodName() got = %v, want = %v", got, tt.want)
			}
		})
	}
}

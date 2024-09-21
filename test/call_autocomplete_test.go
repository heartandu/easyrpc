package test

import (
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestCallAutocomplete(t *testing.T) {
	fs := afero.NewCopyOnWriteFs(afero.NewOsFs(), afero.NewMemMapFs())

	protoConf, err := createTempFile(fs, "proto-autocomp.yaml", `
        import_paths:
          - `+importPath+`
        proto_files:
          - `+protoFile+`
    `)
	if err != nil {
		t.Fatalf("failed to create proto files config file: %v", err)
	}

	reflectConf, err := createTempFile(fs, "reflet-autocomp.yaml", `
        address: `+address(insecureSocket)+`
        reflection: true
    `)
	if err != nil {
		t.Fatalf("failed to create reflect config file: %v", err)
	}

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "empty flags",
			args: []string{""},
			want: []string{},
		},
		{
			name: "empty completion",
			args: []string{
				"-i",
				importPath,
				"-p",
				protoFile,
				"",
			},
			want: []string{
				"echo.EchoService.Echo",
				"echo.EchoService.Error",
				"echo.EchoService.ClientStream",
				"echo.EchoService.ServerStream",
				"echo.EchoService.BidiStream",
			},
		},
		{
			name: "empty completion reflection",
			args: []string{
				"-r",
				"-a",
				address(insecureSocket),
				"",
			},
			want: []string{
				"echo.EchoService.Echo",
				"echo.EchoService.Error",
				"echo.EchoService.ClientStream",
				"echo.EchoService.ServerStream",
				"echo.EchoService.BidiStream",
				"grpc.reflection.v1.ServerReflection.ServerReflectionInfo",
				"grpc.reflection.v1alpha.ServerReflection.ServerReflectionInfo",
			},
		},
		{
			name: "empty completion config",
			args: []string{
				"--config",
				protoConf,
				"",
			},
			want: []string{
				"echo.EchoService.Echo",
				"echo.EchoService.Error",
				"echo.EchoService.ClientStream",
				"echo.EchoService.ServerStream",
				"echo.EchoService.BidiStream",
			},
		},
		{
			name: "empty completion reflection config",
			args: []string{
				"--config",
				reflectConf,
				"",
			},
			want: []string{
				"echo.EchoService.Echo",
				"echo.EchoService.Error",
				"echo.EchoService.ClientStream",
				"echo.EchoService.ServerStream",
				"echo.EchoService.BidiStream",
				"grpc.reflection.v1.ServerReflection.ServerReflectionInfo",
				"grpc.reflection.v1alpha.ServerReflection.ServerReflectionInfo",
			},
		},
		{
			name: "partial completion",
			args: []string{
				"-r",
				"-a",
				address(insecureSocket),
				"err",
			},
			want: []string{
				"echo.EchoService.Error",
				"grpc.reflection.v1.ServerReflection.ServerReflectionInfo",
				"grpc.reflection.v1alpha.ServerReflection.ServerReflectionInfo",
			},
		},
		{
			name: "partial case sensitive completion",
			args: []string{
				"-r",
				"-a",
				address(insecureSocket),
				"Err",
			},
			want: []string{
				"echo.EchoService.Error",
			},
		},
		{
			name: "partial completion over web reflection",
			args: []string{
				"-r",
				"-a",
				address(insecureWebSocket),
				"-w",
				"err",
			},
			want: []string{
				"echo.EchoService.Error",
				"grpc.reflection.v1.ServerReflection.ServerReflectionInfo",
				"grpc.reflection.v1alpha.ServerReflection.ServerReflectionInfo",
			},
		},
		{
			name: "partial completion over tls web reflection",
			args: []string{
				"-r",
				"-a",
				address(tlsWebSocket),
				"--tls",
				"--cacert",
				cacert,
				"--cert",
				cert,
				"--key",
				key,
				"-w",
				"err",
			},
			want: []string{
				"echo.EchoService.Error",
				"grpc.reflection.v1.ServerReflection.ServerReflectionInfo",
				"grpc.reflection.v1alpha.ServerReflection.ServerReflectionInfo",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := runCallAutocomplete(fs, tt.args...)
			if err != nil {
				t.Fatalf("command failed: output = %v, err = %v", string(b), err)
			}

			lines := strings.Split(strings.TrimSpace(string(b)), "\n")
			if len(lines) < 2 {
				t.Fatalf("autocomplete returned unknown response: %v", lines)
			}

			require.Equal(t, tt.want, lines[:len(lines)-2])
		})
	}
}

func runCallAutocomplete(fs afero.Fs, args ...string) ([]byte, error) {
	return run(fs, nil, append([]string{"__complete", "call"}, args...)...)
}

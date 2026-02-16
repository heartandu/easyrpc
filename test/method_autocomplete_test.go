package test

import (
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestMethodAutocomplete(t *testing.T) {
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

	protoImportAllConf, err := createTempFile(fs, "proto-import-all-autocomp.yaml", `
        import_paths:
          - `+importPath+`
        import_all: true`)
	if err != nil {
		t.Fatalf("failed to create proto import all files config file: %v", err)
	}

	reflectConf, err := createTempFile(fs, "reflect-autocomp.yaml", `
        address: `+address(insecureSocket)+`
        reflection: true
    `)
	if err != nil {
		t.Fatalf("failed to create reflect config file: %v", err)
	}

	testCases := []struct {
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
			name: "empty completion import all",
			args: []string{
				"-i",
				importPath,
				"--import-all",
				"",
			},
			want: []string{
				"echo.EchoService.Echo",
				"echo.EchoService.Error",
				"echo.EchoService.ClientStream",
				"echo.EchoService.ServerStream",
				"echo.EchoService.BidiStream",
				"types.TypesService.ScalarTypes",
				"types.TypesService.EnumTypes",
				"types.TypesService.Maps",
				"types.TypesService.Oneof",
				"types.TypesService.Imported",
				"types.TypesService.Recursive",
				"types.TypesService.Optional",
				"types.TypesService.Repeated",
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
				"types.TypesService.ScalarTypes",
				"types.TypesService.EnumTypes",
				"types.TypesService.Maps",
				"types.TypesService.Oneof",
				"types.TypesService.Imported",
				"types.TypesService.Recursive",
				"types.TypesService.Optional",
				"types.TypesService.Repeated",
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
			name: "empty completion config import all",
			args: []string{
				"--config",
				protoImportAllConf,
				"",
			},
			want: []string{
				"echo.EchoService.Echo",
				"echo.EchoService.Error",
				"echo.EchoService.ClientStream",
				"echo.EchoService.ServerStream",
				"echo.EchoService.BidiStream",
				"types.TypesService.ScalarTypes",
				"types.TypesService.EnumTypes",
				"types.TypesService.Maps",
				"types.TypesService.Oneof",
				"types.TypesService.Imported",
				"types.TypesService.Recursive",
				"types.TypesService.Optional",
				"types.TypesService.Repeated",
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
				"types.TypesService.ScalarTypes",
				"types.TypesService.EnumTypes",
				"types.TypesService.Maps",
				"types.TypesService.Oneof",
				"types.TypesService.Imported",
				"types.TypesService.Recursive",
				"types.TypesService.Optional",
				"types.TypesService.Repeated",
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
		{
			name: "empty completion with package flag",
			args: []string{
				"--config",
				reflectConf,
				"--package",
				"echo",
				"",
			},
			want: []string{
				"EchoService.Echo",
				"EchoService.Error",
				"EchoService.ClientStream",
				"EchoService.ServerStream",
				"EchoService.BidiStream",
			},
		},
		{
			name: "partial completion with package flag",
			args: []string{
				"--config",
				reflectConf,
				"--package",
				"echo",
				"stream",
			},
			want: []string{
				"EchoService.ClientStream",
				"EchoService.ServerStream",
				"EchoService.BidiStream",
			},
		},
		{
			name: "empty completion with package and service flags",
			args: []string{
				"--config",
				reflectConf,
				"--package",
				"echo",
				"--service",
				"EchoService",
				"",
			},
			want: []string{
				"Echo",
				"Error",
				"ClientStream",
				"ServerStream",
				"BidiStream",
			},
		},
		{
			name: "partial completion with package and service flags",
			args: []string{
				"--config",
				reflectConf,
				"--package",
				"echo",
				"--service",
				"EchoService",
				"stream",
			},
			want: []string{
				"ClientStream",
				"ServerStream",
				"BidiStream",
			},
		},
		{
			name: "partial completion with only service flag",
			args: []string{
				"--config",
				reflectConf,
				"--service",
				"EchoService",
				"stream",
			},
			want: []string{
				"echo.EchoService.ClientStream",
				"echo.EchoService.ServerStream",
				"echo.EchoService.BidiStream",
			},
		},
		{
			name: "partial completion with fully qualified service",
			args: []string{
				"--config",
				reflectConf,
				"--service",
				"echo.EchoService",
				"stream",
			},
			want: []string{
				"echo.EchoService.ClientStream",
				"echo.EchoService.ServerStream",
				"echo.EchoService.BidiStream",
			},
		},
		{
			name: "partial completion with fully qualified service that doesn't exist",
			args: []string{
				"--config",
				reflectConf,
				"--service",
				"test.EchoService",
				"stream",
			},
			want: []string{},
		},
		{
			name: "partial completion with fully qualified service and package",
			args: []string{
				"--config",
				reflectConf,
				"--package",
				"echo",
				"--service",
				"echo.EchoService",
				"stream",
			},
			want: []string{
				"ClientStream",
				"ServerStream",
				"BidiStream",
			},
		},
	}

	commands := []string{"call", "request"}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			for _, tt := range testCases {
				t.Run(tt.name, func(t *testing.T) {
					b, err := runMethodAutocomplete(fs, cmd, tt.args...)
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
		})
	}
}

func runMethodAutocomplete(fs afero.Fs, command string, args ...string) ([]byte, error) {
	return run(fs, nil, nil, append([]string{"__complete", command}, args...)...)
}

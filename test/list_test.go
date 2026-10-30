package test

import (
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	fs := afero.NewCopyOnWriteFs(afero.NewOsFs(), afero.NewMemMapFs())

	protoConfigFileName, err := createTempFile(fs, "proto.yaml", `
        import_paths:
          - `+importPath+`
        proto_files:
          - `+protoFile)
	require.NoError(t, err, "failed to create proto config file")

	protoImportAllConfigFileName, err := createTempFile(fs, "proto_import_all.yaml", `
        import_paths:
          - `+importPath+`
        import_all: true`)
	require.NoError(t, err, "failed to create proto import all config file")

	reflectionConfigFileName, err := createTempFile(fs, "reflect.yaml", `
        address: `+address(insecureSocket)+`
        reflection: true`)
	require.NoError(t, err, "failed to create reflection config file")

	packageAndServiceConfigFileName, err := createTempFile(fs, "pns.yaml", `
        address: `+address(insecureSocket)+`
        reflection: true
        package: echo
        service: EchoService`)
	require.NoError(t, err, "failed to create package and service config file")

	mixedCasingConfigFileName, err := createTempFile(fs, "mixed_casing.yaml", `
        AddreSS: `+address(insecureSocket)+`
        import_Paths:
          - `+importPath+`
        IMPORT_ALL: true`)
	require.NoError(t, err, "failed to create mixed casing config file")

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "by proto",
			args: []string{
				"-i",
				importPath,
				"-p",
				protoFile,
			},
			want: []string{
				"echo.EchoService.BidiStream",
				"echo.EchoService.ClientStream",
				"echo.EchoService.Echo",
				"echo.EchoService.Error",
				"echo.EchoService.ServerStream",
			},
		},
		{
			name: "by proto with import all",
			args: []string{
				"-i",
				importPath,
				"--import-all",
			},
			want: []string{
				"EchoService.Echo",
				"TimeService.Now",
				"echo.EchoService.BidiStream",
				"echo.EchoService.ClientStream",
				"echo.EchoService.Echo",
				"echo.EchoService.Error",
				"echo.EchoService.ServerStream",
				"types.TypesService.EnumTypes",
				"types.TypesService.Imported",
				"types.TypesService.Maps",
				"types.TypesService.Oneof",
				"types.TypesService.Optional",
				"types.TypesService.Recursive",
				"types.TypesService.Repeated",
				"types.TypesService.ScalarTypes",
			},
		},
		{
			name: "by reflection",
			args: []string{
				"-a",
				address(insecureSocket),
				"-r",
			},
			want: []string{
				"EchoService.Echo",
				"TimeService.Now",
				"echo.EchoService.BidiStream",
				"echo.EchoService.ClientStream",
				"echo.EchoService.Echo",
				"echo.EchoService.Error",
				"echo.EchoService.ServerStream",
				"grpc.reflection.v1.ServerReflection.ServerReflectionInfo",
				"grpc.reflection.v1alpha.ServerReflection.ServerReflectionInfo",
				"types.TypesService.EnumTypes",
				"types.TypesService.Imported",
				"types.TypesService.Maps",
				"types.TypesService.Oneof",
				"types.TypesService.Optional",
				"types.TypesService.Recursive",
				"types.TypesService.Repeated",
				"types.TypesService.ScalarTypes",
			},
		},
		{
			name: "by proto with config",
			args: []string{
				"--config",
				protoConfigFileName,
			},
			want: []string{
				"echo.EchoService.BidiStream",
				"echo.EchoService.ClientStream",
				"echo.EchoService.Echo",
				"echo.EchoService.Error",
				"echo.EchoService.ServerStream",
			},
		},
		{
			name: "by proto import all with config",
			args: []string{
				"--config",
				protoImportAllConfigFileName,
			},
			want: []string{
				"EchoService.Echo",
				"TimeService.Now",
				"echo.EchoService.BidiStream",
				"echo.EchoService.ClientStream",
				"echo.EchoService.Echo",
				"echo.EchoService.Error",
				"echo.EchoService.ServerStream",
				"types.TypesService.EnumTypes",
				"types.TypesService.Imported",
				"types.TypesService.Maps",
				"types.TypesService.Oneof",
				"types.TypesService.Optional",
				"types.TypesService.Recursive",
				"types.TypesService.Repeated",
				"types.TypesService.ScalarTypes",
			},
		},
		{
			name: "by reflection with config",
			args: []string{
				"--config",
				reflectionConfigFileName,
			},
			want: []string{
				"EchoService.Echo",
				"TimeService.Now",
				"echo.EchoService.BidiStream",
				"echo.EchoService.ClientStream",
				"echo.EchoService.Echo",
				"echo.EchoService.Error",
				"echo.EchoService.ServerStream",
				"grpc.reflection.v1.ServerReflection.ServerReflectionInfo",
				"grpc.reflection.v1alpha.ServerReflection.ServerReflectionInfo",
				"types.TypesService.EnumTypes",
				"types.TypesService.Imported",
				"types.TypesService.Maps",
				"types.TypesService.Oneof",
				"types.TypesService.Optional",
				"types.TypesService.Recursive",
				"types.TypesService.Repeated",
				"types.TypesService.ScalarTypes",
			},
		},
		{
			name: "using mixed casing keys in config",
			args: []string{
				"--config",
				mixedCasingConfigFileName,
			},
			want: []string{
				"EchoService.Echo",
				"TimeService.Now",
				"echo.EchoService.BidiStream",
				"echo.EchoService.ClientStream",
				"echo.EchoService.Echo",
				"echo.EchoService.Error",
				"echo.EchoService.ServerStream",
				"types.TypesService.EnumTypes",
				"types.TypesService.Imported",
				"types.TypesService.Maps",
				"types.TypesService.Oneof",
				"types.TypesService.Optional",
				"types.TypesService.Recursive",
				"types.TypesService.Repeated",
				"types.TypesService.ScalarTypes",
			},
		},
		{
			name: "package flag specified",
			args: []string{
				"-a",
				address(insecureSocket),
				"-r",
				"--package",
				"echo",
			},
			want: []string{
				"EchoService.BidiStream",
				"EchoService.ClientStream",
				"EchoService.Echo",
				"EchoService.Error",
				"EchoService.ServerStream",
			},
		},
		{
			name: "service flag specified",
			args: []string{
				"-a",
				address(insecureSocket),
				"-r",
				"--service",
				"EchoService",
			},
			want: []string{
				"Echo",
				"echo.EchoService.BidiStream",
				"echo.EchoService.ClientStream",
				"echo.EchoService.Echo",
				"echo.EchoService.Error",
				"echo.EchoService.ServerStream",
			},
		},
		{
			name: "package and service flags specified",
			args: []string{
				"-a",
				address(insecureSocket),
				"-r",
				"--package",
				"echo",
				"--service",
				"EchoService",
			},
			want: []string{
				"BidiStream",
				"ClientStream",
				"Echo",
				"Error",
				"ServerStream",
			},
		},
		{
			name: "package and service config file specified",
			args: []string{
				"--config",
				packageAndServiceConfigFileName,
			},
			want: []string{
				"BidiStream",
				"ClientStream",
				"Echo",
				"Error",
				"ServerStream",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := runList(fs, tt.args...)
			require.NoErrorf(t, err, "command failed with output: %s", string(b))

			got := strings.Split(strings.TrimSpace(string(b)), "\n")
			if len(got) == 1 && got[0] == "" {
				got = nil
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestList_ErrorCases(t *testing.T) {
	fs := afero.NewCopyOnWriteFs(afero.NewOsFs(), afero.NewMemMapFs())

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name: "unexpected arguments",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				address(insecureSocket),
				"-r",
			},
			wantErr: "unexpected arguments",
		},
		{
			name: "missing reflection and proto files",
			args: []string{
				"-a",
				address(insecureSocket),
			},
			wantErr: "at least 1 proto file must be specified, imported all files or reflection used",
		},
		{
			name: "connection refused",
			args: []string{
				"-a",
				"localhost:59999",
				"-r",
			},
			wantErr: "connection refused",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := runList(fs, tt.args...)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func runList(fs afero.Fs, args ...string) ([]byte, error) {
	return run(fs, nil, nil, append([]string{"list"}, args...)...)
}

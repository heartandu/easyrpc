package test

import (
	"iter"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestCallAutocomplete(t *testing.T) {
	fs := afero.NewMemMapFs()

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
        address: `+insecureAddress+`
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
			want: nil,
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
			},
		},
		{
			name: "empty completion reflection",
			args: []string{
				"-r",
				"-a",
				insecureAddress,
				"",
			},
			want: []string{
				"echo.EchoService.Echo",
				"echo.EchoService.Error",
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
				"grpc.reflection.v1.ServerReflection.ServerReflectionInfo",
				"grpc.reflection.v1alpha.ServerReflection.ServerReflectionInfo",
			},
		},
		{
			name: "partial completion",
			args: []string{
				"-r",
				"-a",
				insecureAddress,
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
				insecureAddress,
				"Err",
			},
			want: []string{
				"echo.EchoService.Error",
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

			require.Equal(t, tt.want, slices.Collect(dedup(lines[:len(lines)-2])))
		})
	}
}

func runCallAutocomplete(fs afero.Fs, args ...string) ([]byte, error) {
	return run(fs, nil, append([]string{"__complete", "call"}, args...)...)
}

func dedup(lines []string) iter.Seq[string] {
	return func(yield func(string) bool) {
		if len(lines) == 0 {
			return
		}

		m := map[string]struct{}{}

		for _, l := range lines {
			if _, ok := m[l]; !ok {
				m[l] = struct{}{}

				if cont := yield(l); !cont {
					return
				}
			}
		}
	}
}

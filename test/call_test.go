package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// TODO: Add streaming call tests, add more protobuf types to tests.
func TestCall(t *testing.T) {
	fs := afero.NewMemMapFs()

	requestFileName, err := createTempFile(fs, "msg.json", `{"msg":"file test"}`)
	if err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	protoConfigFileName, err := createTempFile(fs, "proto.yaml", `
        address: `+insecureAddress+`
        import_paths:
          - `+importPath+`
        proto_files:
          - `+protoFile)
	if err != nil {
		t.Fatalf("failed to create proto config file: %v", err)
	}

	reflectionConfigFileName, err := createTempFile(fs, "reflect.yaml", `
        address: `+insecureAddress+`
        reflection: true`)
	if err != nil {
		t.Fatalf("failed to create proto config file: %v", err)
	}

	tlsConfigFileName, err := createTempFile(fs, "tls.yaml", `
        address: `+tlsAddress+`
        reflection: true
        tls: true
        cacert: `+cacert+`
        cert: `+cert+`
        cert_key: `+certKey)
	if err != nil {
		t.Fatalf("failed to create tls config file: %v", err)
	}

	packageAndServiceConfigFileName, err := createTempFile(fs, "pns.yaml", `
        address: `+insecureAddress+`
        reflection: true
        package: echo
        service: EchoService`)
	if err != nil {
		t.Fatalf("failed to create proto config file: %v", err)
	}

	mdConfigFileName, err := createTempFile(fs, "md.yaml", `
        address: `+insecureAddress+`
        reflection: true
        metadata:
          test: config`)
	if err != nil {
		t.Fatalf("failed to create metadata config file: %v", err)
	}

	tests := []struct {
		name string
		args []string
		in   io.Reader
		want []map[string]any
	}{
		{
			name: "by proto",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				insecureAddress,
				"-d",
				`{"msg":"oops"}`,
				"-i",
				importPath,
				"-p",
				protoFile,
			},
			want: []map[string]any{{"msg": "oops"}},
		},
		{
			name: "by reflection",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				insecureAddress,
				"-d",
				`{"msg":"hello"}`,
				"-r",
			},
			want: []map[string]any{{"msg": "hello"}},
		},
		{
			name: "data from file",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				insecureAddress,
				"-d",
				"@" + requestFileName,
				"-r",
			},
			want: []map[string]any{{"msg": "file test"}},
		},
		{
			name: "data from stdin",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				insecureAddress,
				"-d",
				"@",
				"-r",
			},
			in:   strings.NewReader(`{"msg":"stdin test"}`),
			want: []map[string]any{{"msg": "stdin test"}},
		},
		{
			name: "by proto with config",
			args: []string{
				"echo.EchoService.Echo",
				"--config",
				protoConfigFileName,
				"-d",
				`{"msg":"proto config"}`,
			},
			want: []map[string]any{{"msg": "proto config"}},
		},
		{
			name: "by reflection with config",
			args: []string{
				"echo.EchoService.Echo",
				"--config",
				reflectionConfigFileName,
				"-d",
				`{"msg":"reflection config"}`,
			},
			want: []map[string]any{{"msg": "reflection config"}},
		},
		{
			name: "tls with only root certificate",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				tlsAddress,
				"-d",
				`{"msg":"tls"}`,
				"-r",
				"--tls",
				"--cacert",
				cacert,
			},
			want: []map[string]any{{"msg": "tls"}},
		},
		{
			name: "tls with server certificates",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				tlsAddress,
				"-d",
				`{"msg":"tls certs"}`,
				"-r",
				"--tls",
				"--cacert",
				cacert,
				"--cert",
				cert,
				"--cert-key",
				certKey,
			},
			want: []map[string]any{{"msg": "tls certs"}},
		},
		{
			name: "tls with server certificates config",
			args: []string{
				"echo.EchoService.Echo",
				"--config",
				tlsConfigFileName,
				"-d",
				`{"msg":"tls certs config"}`,
			},
			want: []map[string]any{{"msg": "tls certs config"}},
		},
		{
			name: "package flag specified",
			args: []string{
				"EchoService.Echo",
				"-a",
				insecureAddress,
				"-d",
				`{"msg":"package flag"}`,
				"-r",
				"--package",
				"echo",
			},
			want: []map[string]any{{"msg": "package flag"}},
		},
		{
			name: "package and service flag specified",
			args: []string{
				"Echo",
				"-a",
				insecureAddress,
				"-d",
				`{"msg":"package and service flags"}`,
				"-r",
				"--package",
				"echo",
				"--service",
				"EchoService",
			},
			want: []map[string]any{{"msg": "package and service flags"}},
		},
		{
			name: "package and service config file specified",
			args: []string{
				"Echo",
				"--config",
				packageAndServiceConfigFileName,
				"-d",
				`{"msg":"package and service flags"}`,
			},
			want: []map[string]any{{"msg": "package and service flags"}},
		},
		{
			name: "with metadata flag",
			args: []string{
				"echo.EchoService.Echo",
				"-r",
				"-a",
				insecureAddress,
				"-d",
				`{"msg":"md flag"}`,
				"-H",
				"test=test",
			},
			want: []map[string]any{{"msg": "md flag\ntest"}},
		},
		{
			name: "with metadata in config",
			args: []string{
				"echo.EchoService.Echo",
				"--config",
				mdConfigFileName,
				"-d",
				`{"msg":"md flag"}`,
			},
			want: []map[string]any{{"msg": "md flag\nconfig"}},
		},
		{
			name: "with metadata flag precedence",
			args: []string{
				"echo.EchoService.Echo",
				"--config",
				mdConfigFileName,
				"-d",
				`{"msg":"md flag"}`,
				"-H",
				"test=overwritten",
			},
			want: []map[string]any{{"msg": "md flag\noverwritten"}},
		},
		{
			name: "client streaming request",
			args: []string{
				"echo.EchoService.ClientStream",
				"-r",
				"-a",
				insecureAddress,
				"-d",
				`{"msg":"1"}{"msg":"3"}{"msg":"2"}`,
				"-H",
				"test=321",
			},
			want: []map[string]any{{"msgs": []any{"1", "3", "2", "321"}}},
		},
		{
			name: "server streaming request",
			args: []string{
				"echo.EchoService.ServerStream",
				"-r",
				"-a",
				insecureAddress,
				"-d",
				`{"msgs":["1", "3", "2"]}`,
				"-H",
				"test=321",
			},
			want: []map[string]any{{"msg": "1"}, {"msg": "3"}, {"msg": "2"}, {"msg": "321"}},
		},
		{
			name: "bidi streaming request",
			args: []string{
				"echo.EchoService.BidiStream",
				"-r",
				"-a",
				insecureAddress,
				"-d",
				`{"msg":"1"}{"msg":"3"}{"msg":"2"}`,
				"-H",
				"test=321",
			},
			want: []map[string]any{{"msg": "1"}, {"msg": "3"}, {"msg": "2"}, {"msg": "321"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := runCall(fs, tt.in, tt.args...)
			if err != nil {
				t.Fatalf("command failed: output = %v, err = %v", string(b), err)
			}

			got := []map[string]any{}
			d := json.NewDecoder(bytes.NewReader(b))

			for {
				v := map[string]any{}
				if err := d.Decode(&v); err != nil {
					if errors.Is(err, io.EOF) {
						break
					}

					t.Fatalf("failed to decode output: %v", err)
				}

				got = append(got, v)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func runCall(fs afero.Fs, in io.Reader, args ...string) ([]byte, error) {
	return run(fs, in, append([]string{"call"}, args...)...)
}

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

// TODO: Add more protobuf types to tests.
func TestCall(t *testing.T) {
	fs := afero.NewCopyOnWriteFs(afero.NewOsFs(), afero.NewMemMapFs())

	requestFileName, err := createTempFile(fs, "msg.json", `{"msg":"file test"}`)
	require.NoError(t, err, "failed to create input file")

	reqWithUnknownFileName, err := createTempFile(
		fs,
		"dirty_msg.json",
		`{"$schema":"https://example.com/some/schema.json","msg":"file with unknown"}`,
	)
	require.NoError(t, err, "failed to create input file with unknown fields")

	protoConfigFileName, err := createTempFile(fs, "proto.yaml", `
        address: `+address(insecureSocket)+`
        import_paths:
          - `+importPath+`
        proto_files:
          - `+protoFile)
	require.NoError(t, err, "failed to create proto config file")

	protoImportAllConfigFileName, err := createTempFile(fs, "proto_import_all.yaml", `
        address: `+address(insecureSocket)+`
        import_paths:
          - `+importPath+`
        import_all: true`)
	require.NoError(t, err, "failed to create proto import all config file")

	reflectionConfigFileName, err := createTempFile(fs, "reflect.yaml", `
        address: `+address(insecureSocket)+`
        reflection: true`)
	require.NoError(t, err, "failed to create proto reflection config file")

	tlsConfigFileName, err := createTempFile(fs, "tls.yaml", `
        address: `+address(tlsSocket)+`
        reflection: true
        tls: true
        cacert: `+cacert+`
        cert: `+cert+`
        key: `+key)
	require.NoError(t, err, "failed to create tls config file")

	packageAndServiceConfigFileName, err := createTempFile(fs, "pns.yaml", `
        address: `+address(insecureSocket)+`
        reflection: true
        package: echo
        service: EchoService`)
	require.NoError(t, err, "failed to create proto config file")

	mdConfigFileName, err := createTempFile(fs, "md.yaml", `
        address: `+address(insecureSocket)+`
        reflection: true
        metadata:
          test: config`)
	require.NoError(t, err, "failed to create metadata config file")

	webConfigFileName, err := createTempFile(fs, "web.yaml", `
        address: `+address(insecureWebSocket)+`
        reflection: true
        web: true`)
	require.NoError(t, err, "failed to create metadata config file")

	webTLSConfigFileName, err := createTempFile(fs, "webTLS.yaml", `
        address: `+address(tlsWebSocket)+`
        cacert: `+cacert+`
        cert: `+cert+`
        key: `+key+`
        reflection: true
        tls: true
        web: true`)
	require.NoError(t, err, "failed to create metadata config file")

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
				address(insecureSocket),
				"-d",
				`{"msg":"oops"}`,
				"-i",
				importPath,
				"-p",
				protoFile,
			},
			want: []map[string]any{{"msg": "oops"}},
		},
		// TODO: add more tests with more deeply nested proto files and multiple import paths
		{
			name: "by proto with import all",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				address(insecureSocket),
				"-d",
				`{"msg":"good"}`,
				"-i",
				importPath,
				"--import-all",
			},
			want: []map[string]any{{"msg": "good"}},
		},
		{
			name: "by reflection",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				address(insecureSocket),
				"-d",
				`{"msg":"hello"}`,
				"-r",
			},
			want: []map[string]any{{"msg": "hello"}},
		},

		{
			name: "data from flag with unknown field",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				address(insecureSocket),
				"-d",
				`{"$schema":"https://example.com/some/schema.json","msg":"with unknown"}`,
				"-r",
			},
			want: []map[string]any{{"msg": "with unknown"}},
		},
		{
			name: "data from file",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				address(insecureSocket),
				"-d",
				"@" + requestFileName,
				"-r",
			},
			want: []map[string]any{{"msg": "file test"}},
		},
		{
			name: "data from file with unknown field",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				address(insecureSocket),
				"-d",
				"@" + reqWithUnknownFileName,
				"-r",
			},
			want: []map[string]any{{"msg": "file with unknown"}},
		},
		{
			name: "data from stdin",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				address(insecureSocket),
				"-d",
				"-",
				"-r",
			},
			in:   strings.NewReader(`{"msg":"stdin test"}`),
			want: []map[string]any{{"msg": "stdin test"}},
		},
		{
			name: "data from stdin with unknown field",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				address(insecureSocket),
				"-d",
				"-",
				"-r",
			},
			in:   strings.NewReader(`{"$schema":"https://example.com/some/schema.json","msg":"stdin with unknown"}`),
			want: []map[string]any{{"msg": "stdin with unknown"}},
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
			name: "by proto import all with config",
			args: []string{
				"echo.EchoService.Echo",
				"--config",
				protoImportAllConfigFileName,
				"-d",
				`{"msg":"proto import all"}`,
			},
			want: []map[string]any{{"msg": "proto import all"}},
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
				address(tlsSocket),
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
				address(tlsSocket),
				"-d",
				`{"msg":"tls certs"}`,
				"-r",
				"--tls",
				"--cacert",
				cacert,
				"--cert",
				cert,
				"--key",
				key,
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
				address(insecureSocket),
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
				address(insecureSocket),
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
				address(insecureSocket),
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
				address(insecureSocket),
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
				address(insecureSocket),
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
				address(insecureSocket),
				"-d",
				`{"msg":"1"}{"msg":"3"}{"msg":"2"}`,
				"-H",
				"test=321",
			},
			want: []map[string]any{{"msg": "1"}, {"msg": "3"}, {"msg": "2"}, {"msg": "321"}},
		},
		{
			name: "web unary request with config",
			args: []string{
				"echo.EchoService.Echo",
				"--config",
				webConfigFileName,
				"-d",
				`{"msg":"web config"}`,
			},
			want: []map[string]any{{"msg": "web config"}},
		},
		{
			name: "web unary request",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				address(insecureWebSocket),
				"-w",
				"-i",
				importPath,
				"-p",
				protoFile,
				"-d",
				`{"msg":"web unary"}`,
			},
			want: []map[string]any{{"msg": "web unary"}},
		},
		{
			name: "web client streaming request",
			args: []string{
				"echo.EchoService.ClientStream",
				"--config",
				webTLSConfigFileName,
				"-H",
				"test=321",
				"-d",
				`{"msg":"1"}{"msg":"3"}{"msg":"2"}`,
			},
			want: []map[string]any{{"msgs": []any{"1", "3", "2", "321"}}},
		},
		{
			name: "web server streaming request",
			args: []string{
				"echo.EchoService.ServerStream",
				"--config",
				webTLSConfigFileName,
				"-d",
				`{"msgs":["1", "3", "2"]}`,
				"-H",
				"test=321",
			},
			want: []map[string]any{{"msg": "1"}, {"msg": "3"}, {"msg": "2"}, {"msg": "321"}},
		},
		{
			name: "web bidi streaming request",
			args: []string{
				"echo.EchoService.BidiStream",
				"--config",
				webTLSConfigFileName,
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
			require.NoErrorf(t, err, "command failed with output: %s", string(b))

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

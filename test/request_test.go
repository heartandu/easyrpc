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

func TestRequest(t *testing.T) {
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

	webConfigFileName, err := createTempFile(fs, "web.yaml", `
        address: `+address(insecureWebSocket)+`
        reflection: true
        web: true`)
	require.NoError(t, err, "failed to create web config file")

	webTLSConfigFileName, err := createTempFile(fs, "webTLS.yaml", `
        address: `+address(tlsWebSocket)+`
        cacert: `+cacert+`
        cert: `+cert+`
        key: `+key+`
        reflection: true
        tls: true
        web: true`)
	require.NoError(t, err, "failed to create web TLS config file")

	tests := []struct {
		name string
		args []string
		in   io.Reader
		want map[string]any
	}{
		{
			name: "by proto",
			args: []string{
				"echo.EchoService.Echo",
				"-d",
				`{"msg":"oops"}`,
				"-i",
				importPath,
				"-p",
				protoFile,
			},
			want: map[string]any{"msg": "oops"},
		},
		{
			name: "by proto with import all",
			args: []string{
				"echo.EchoService.Echo",
				"-d",
				`{"msg":"good"}`,
				"-i",
				importPath,
				"--import-all",
			},
			want: map[string]any{"msg": "good"},
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
			want: map[string]any{"msg": "hello"},
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
			want: map[string]any{"msg": "with unknown"},
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
			want: map[string]any{"msg": "file test"},
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
			want: map[string]any{"msg": "file with unknown"},
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
			want: map[string]any{"msg": "stdin test"},
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
			want: map[string]any{"msg": "stdin with unknown"},
		},
		{
			name: "empty data outputs template",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				address(insecureSocket),
				"-r",
			},
			want: map[string]any{"msg": ""},
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
			want: map[string]any{"msg": "proto config"},
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
			want: map[string]any{"msg": "proto import all"},
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
			want: map[string]any{"msg": "reflection config"},
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
			want: map[string]any{"msg": "tls"},
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
			want: map[string]any{"msg": "tls certs"},
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
			want: map[string]any{"msg": "tls certs config"},
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
			want: map[string]any{"msg": "package flag"},
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
			want: map[string]any{"msg": "package and service flags"},
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
			want: map[string]any{"msg": "package and service flags"},
		},
		{
			name: "web request with config",
			args: []string{
				"echo.EchoService.Echo",
				"--config",
				webConfigFileName,
				"-d",
				`{"msg":"web config"}`,
			},
			want: map[string]any{"msg": "web config"},
		},
		{
			name: "web request",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				address(insecureWebSocket),
				"-w",
				"-r",
				"-d",
				`{"msg":"web unary"}`,
			},
			want: map[string]any{"msg": "web unary"},
		},
		{
			name: "web tls request with config",
			args: []string{
				"echo.EchoService.Echo",
				"--config",
				webTLSConfigFileName,
				"-d",
				`{"msg":"web tls config"}`,
			},
			want: map[string]any{"msg": "web tls config"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := runRequest(fs, tt.in, tt.args...)
			require.NoErrorf(t, err, "command failed with output: %s", string(b))

			got := map[string]any{}
			require.NoError(t, json.Unmarshal(b, &got))
			require.Equal(t, tt.want, got)
		})
	}
}

func TestRequest_MultipleMessages(t *testing.T) {
	fs := afero.NewCopyOnWriteFs(afero.NewOsFs(), afero.NewMemMapFs())

	tests := []struct {
		name string
		args []string
		want []map[string]any
	}{
		{
			name: "client streaming request format",
			args: []string{
				"echo.EchoService.ClientStream",
				"-r",
				"-a",
				address(insecureSocket),
				"-d",
				`{"msg":"1"}{"msg":"3"}{"msg":"2"}`,
			},
			want: []map[string]any{{"msg": "1"}, {"msg": "3"}, {"msg": "2"}},
		},
		{
			name: "bidi streaming request format",
			args: []string{
				"echo.EchoService.BidiStream",
				"-r",
				"-a",
				address(insecureSocket),
				"-d",
				`{"msg":"1"}{"msg":"3"}{"msg":"2"}`,
			},
			want: []map[string]any{{"msg": "1"}, {"msg": "3"}, {"msg": "2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := runRequest(fs, nil, tt.args...)
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

func TestRequest_OutputToFile(t *testing.T) {
	const outputFileName = "output.json"

	fs := afero.NewCopyOnWriteFs(afero.NewOsFs(), afero.NewMemMapFs())

	b, err := runRequest(
		fs,
		nil,
		"echo.EchoService.Echo",
		"-a",
		address(insecureSocket),
		"-d",
		`{"msg":"output to file"}`,
		"-r",
		"-o",
		outputFileName,
	)
	require.NoErrorf(t, err, "command failed with output: %s", string(b))

	got := map[string]any{}
	file, err := fs.Open(outputFileName)
	require.NoError(t, err, "failed to open output file")

	defer file.Close()

	require.NoError(t, json.NewDecoder(file).Decode(&got))
	require.Equal(t, map[string]any{"msg": "output to file"}, got)
}

func TestRequest_ErrorCases(t *testing.T) {
	fs := afero.NewCopyOnWriteFs(afero.NewOsFs(), afero.NewMemMapFs())

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name: "invalid method name",
			args: []string{
				"echo.NonExistentService.Echo",
				"-a",
				address(insecureSocket),
				"-d",
				`{"msg":"test"}`,
				"-r",
			},
			wantErr: "Symbol not found: echo.NonExistentService.Echo",
		},
		{
			name: "missing reflection and proto files",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				address(insecureSocket),
				"-d",
				`{"msg":"test"}`,
			},
			wantErr: "at least 1 proto file must be specified, imported all files or reflection used",
		},
		{
			name: "connection refused",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				"localhost:59999",
				"-r",
				"-d",
				`{"msg":"test"}`,
			},
			wantErr: "connection refused",
		},
		{
			name: "invalid json in data flag",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				address(insecureSocket),
				"-d",
				`{invalid json}`,
				"-r",
			},
			wantErr: "invalid character 'i' looking for beginning of object key string",
		},
		{
			name: "file not found",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				address(insecureSocket),
				"-d",
				"@/nonexistent/file.json",
				"-r",
			},
			wantErr: "file does not exist",
		},
		{
			name: "unknown method without package and service flags",
			args: []string{
				"UnknownMethod",
				"-a",
				address(insecureSocket),
				"-r",
			},
			wantErr: "Symbol not found: ..UnknownMethod",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := runRequest(fs, nil, tt.args...)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func runRequest(fs afero.Fs, in io.Reader, args ...string) ([]byte, error) {
	return run(fs, in, nil, append([]string{"request"}, args...)...)
}

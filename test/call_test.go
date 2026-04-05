package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestCall(t *testing.T) {
	const tildeTestFileName = "easyrpc_test_msg.json"

	fs := afero.NewCopyOnWriteFs(afero.NewOsFs(), afero.NewMemMapFs())

	requestFileName, err := createTempFile(fs, "msg.json", `{"msg":"file test"}`)
	require.NoError(t, err, "failed to create input file")

	err = createFileAtPath(fs, filepath.Join(homeDir(t), tildeTestFileName), `{"msg":"tilde test"}`)
	require.NoError(t, err, "failed to create tilde test file")

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
				"-f",
				requestFileName,
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
				"-f",
				reqWithUnknownFileName,
				"-r",
			},
			want: []map[string]any{{"msg": "file with unknown"}},
		},
		{
			name: "data from file with tilde path",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				address(insecureSocket),
				"-f",
				"~/" + tildeTestFileName,
				"-r",
			},
			want: []map[string]any{{"msg": "tilde test"}},
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
			name: "service flag specified (should call /EchoService/Echo)",
			args: []string{
				"Echo",
				"-a",
				address(insecureSocket),
				"-d",
				`{"msg":"packageless service flag"}`,
				"-r",
				"--service",
				"EchoService",
			},
			want: []map[string]any{{"msg": "packageless service flag"}},
		},
		{
			name: "service flag specified (should call /echo.EchoService/Echo)",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				address(insecureSocket),
				"-d",
				`{"msg":"packageless service flag"}`,
				"-r",
				"--service",
				"EchoService",
			},
			want: []map[string]any{{"msg": "packageless service flag"}},
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
				"test: test",
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
				"test: overwritten",
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
				"test: 321",
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
				"test: 321",
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
				"test: 321",
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
				"test: 321",
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
				"test: 321",
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
				"test: 321",
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

func TestCall_ErrorCases(t *testing.T) {
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
				"-f",
				"/nonexistent/file.json",
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
			wantErr: "Symbol not found: UnknownMethod",
		},
		{
			name: "both data and file flags specified",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				address(insecureSocket),
				"-d",
				`{"msg":"test"}`,
				"-f",
				"/some/file.json",
				"-r",
			},
			wantErr: "only data or file flag is allowed to be set",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := runCall(fs, nil, tt.args...)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestCall_Types(t *testing.T) {
	fs := afero.NewCopyOnWriteFs(afero.NewOsFs(), afero.NewMemMapFs())

	typesConfigFileName, err := createTempFile(fs, "types_config.yaml", `
        address: `+address(insecureSocket)+`
        reflection: true`)
	require.NoError(t, err, "failed to create types config file")

	tests := []struct {
		name string
		args []string
		want map[string]any
	}{
		{
			name: "scalar types with edge values",
			args: []string{
				"types.TypesService.ScalarTypes",
				"--config",
				typesConfigFileName,
				"-d",
				`{` +
					`"doubleField":1.79769313486231570814527423731704356798070e+308,` +
					`"floatField":3.40282346638528859811704183484516925440e+38,` +
					`"int32Field":2147483647,` +
					`"int64Field":9223372036854775807,` +
					`"uint32Field":4294967295,` +
					`"uint64Field":18446744073709551615,` +
					`"sint32Field":2147483647,` +
					`"sint64Field":9223372036854775807,` +
					`"fixed32Field":4294967295,` +
					`"fixed64Field":18446744073709551615,` +
					`"sfixed32Field":2147483647,` +
					`"sfixed64Field":9223372036854775807,` +
					`"boolField":true,` +
					`"stringField":"test string",` +
					`"bytesField":"YmFzZTY0IGVuY29kZWQ="` +
					`}`,
			},
			want: map[string]any{
				"doubleField":   1.79769313486231570814527423731704356798070e+308,
				"floatField":    3.4028235e+38,
				"int32Field":    2147483647.0,
				"int64Field":    "9223372036854775807",
				"uint32Field":   4294967295.0,
				"uint64Field":   "18446744073709551615",
				"sint32Field":   2147483647.0,
				"sint64Field":   "9223372036854775807",
				"fixed32Field":  4294967295.0,
				"fixed64Field":  "18446744073709551615",
				"sfixed32Field": 2147483647.0,
				"sfixed64Field": "9223372036854775807",
				"boolField":     true,
				"stringField":   "test string",
				"bytesField":    "YmFzZTY0IGVuY29kZWQ=",
			},
		},
		{
			name: "scalar types with zero values",
			args: []string{
				"types.TypesService.ScalarTypes",
				"--config",
				typesConfigFileName,
				"-d",
				`{` +
					`"doubleField":0,` +
					`"floatField":0,` +
					`"int32Field":0,` +
					`"int64Field":0,` +
					`"uint32Field":0,` +
					`"uint64Field":0,` +
					`"sint32Field":0,` +
					`"sint64Field":0,` +
					`"fixed32Field":0,` +
					`"fixed64Field":0,` +
					`"sfixed32Field":0,` +
					`"sfixed64Field":0,` +
					`"boolField":false,` +
					`"stringField":"",` +
					`"bytesField":""` +
					`}`,
			},
			want: map[string]any{
				"doubleField":   0.0,
				"floatField":    0.0,
				"int32Field":    0.0,
				"int64Field":    "0",
				"uint32Field":   0.0,
				"uint64Field":   "0",
				"sint32Field":   0.0,
				"sint64Field":   "0",
				"fixed32Field":  0.0,
				"fixed64Field":  "0",
				"sfixed32Field": 0.0,
				"sfixed64Field": "0",
				"boolField":     false,
				"stringField":   "",
				"bytesField":    "",
			},
		},
		{
			name: "scalar types with negative values",
			args: []string{
				"types.TypesService.ScalarTypes",
				"--config",
				typesConfigFileName,
				"-d",
				`{` +
					`"doubleField":-1.7976931348623157e+308,` +
					`"floatField":-3.4028235e+38,` +
					`"int32Field":-2147483648,` +
					`"int64Field":-9223372036854775808,` +
					`"sint32Field":-2147483648,` +
					`"sint64Field":-9223372036854775808,` +
					`"sfixed32Field":-2147483648,` +
					`"sfixed64Field":-9223372036854775808` +
					`}`,
			},
			want: map[string]any{
				"doubleField":   -1.7976931348623157e+308,
				"floatField":    -3.4028235e+38,
				"int32Field":    -2147483648.0,
				"int64Field":    "-9223372036854775808",
				"uint32Field":   0.0,
				"uint64Field":   "0",
				"sint32Field":   -2147483648.0,
				"sint64Field":   "-9223372036854775808",
				"fixed32Field":  0.0,
				"fixed64Field":  "0",
				"sfixed32Field": -2147483648.0,
				"sfixed64Field": "-9223372036854775808",
				"boolField":     false,
				"stringField":   "",
				"bytesField":    "",
			},
		},
		{
			name: "scalar types with special float values",
			args: []string{
				"types.TypesService.ScalarTypes",
				"--config",
				typesConfigFileName,
				"-d",
				`{"doubleField":"NaN","floatField":"Infinity"}`,
			},
			want: map[string]any{
				"doubleField":   "NaN",
				"floatField":    "Infinity",
				"int32Field":    0.0,
				"int64Field":    "0",
				"uint32Field":   0.0,
				"uint64Field":   "0",
				"sint32Field":   0.0,
				"sint64Field":   "0",
				"fixed32Field":  0.0,
				"fixed64Field":  "0",
				"sfixed32Field": 0.0,
				"sfixed64Field": "0",
				"boolField":     false,
				"stringField":   "",
				"bytesField":    "",
			},
		},
		{
			name: "enum types",
			args: []string{
				"types.TypesService.EnumTypes",
				"--config",
				typesConfigFileName,
				"-d",
				`{"status":"PENDING","priority":"HIGH"}`,
			},
			want: map[string]any{"status": "PENDING", "priority": "HIGH"},
		},
		{
			name: "enum types with numeric values",
			args: []string{
				"types.TypesService.EnumTypes",
				"--config",
				typesConfigFileName,
				"-d",
				`{"status":2,"priority":3}`,
			},
			want: map[string]any{"status": "RUNNING", "priority": "CRITICAL"},
		},
		{
			name: "enum types with default values",
			args: []string{
				"types.TypesService.EnumTypes",
				"--config",
				typesConfigFileName,
				"-d",
				`{}`,
			},
			want: map[string]any{"status": "UNKNOWN", "priority": "LOW"},
		},
		{
			name: "maps",
			args: []string{
				"types.TypesService.Maps",
				"--config",
				typesConfigFileName,
				"-d",
				`{` +
					`"stringToInt":{"key1":100,"key2":200},` +
					`"intToString":{"123":"value1","456":"value2"},` +
					`"stringToMessage":{"msg1":{"id":5}},` +
					`"boolToFloat":{"true":3.14,"false":2.71}` +
					`}`,
			},
			want: map[string]any{
				"stringToInt": map[string]any{"key1": 100.0, "key2": 200.0},
				"intToString": map[string]any{"123": "value1", "456": "value2"},
				"stringToMessage": map[string]any{
					"msg1": map[string]any{"id": "5"},
				},
				"boolToFloat": map[string]any{"true": 3.14, "false": 2.71},
			},
		},
		{
			name: "maps empty",
			args: []string{
				"types.TypesService.Maps",
				"--config",
				typesConfigFileName,
				"-d",
				`{}`,
			},
			want: map[string]any{
				"stringToInt":     map[string]any{},
				"intToString":     map[string]any{},
				"stringToMessage": map[string]any{},
				"boolToFloat":     map[string]any{},
			},
		},
		{
			name: "oneof with string",
			args: []string{
				"types.TypesService.Oneof",
				"--config",
				typesConfigFileName,
				"-d",
				`{"text":"hello world"}`,
			},
			want: map[string]any{"text": "hello world"},
		},
		{
			name: "oneof with number",
			args: []string{
				"types.TypesService.Oneof",
				"--config",
				typesConfigFileName,
				"-d",
				`{"number":42}`,
			},
			want: map[string]any{"number": 42.0},
		},
		{
			name: "oneof with bool",
			args: []string{
				"types.TypesService.Oneof",
				"--config",
				typesConfigFileName,
				"-d",
				`{"flag":true}`,
			},
			want: map[string]any{"flag": true},
		},
		{
			name: "oneof with inner message",
			args: []string{
				"types.TypesService.Oneof",
				"--config",
				typesConfigFileName,
				"-d",
				`{"inner":{"id":99}}`,
			},
			want: map[string]any{
				"inner": map[string]any{"id": "99"},
			},
		},
		{
			name: "imported message",
			args: []string{
				"types.TypesService.Imported",
				"--config",
				typesConfigFileName,
				"-d",
				`{"imported":{"id":"imported-id"}}`,
			},
			want: map[string]any{
				"imported": map[string]any{"id": "imported-id"},
			},
		},
		{
			name: "recursive messages with 2 levels",
			args: []string{
				"types.TypesService.Recursive",
				"--config",
				typesConfigFileName,
				"-d",
				`{"node":{"value":"level0","child":{"value":"level1","child":{"value":"level2"},"children":[{"value":"child1"},{"value":"child2"}]}}}`,
			},
			want: map[string]any{
				"node": map[string]any{
					"value": "level0",
					"child": map[string]any{
						"value": "level1",
						"child": map[string]any{
							"value":    "level2",
							"child":    nil,
							"children": []any{},
						},
						"children": []any{
							map[string]any{"value": "child1", "child": nil, "children": []any{}},
							map[string]any{"value": "child2", "child": nil, "children": []any{}},
						},
					},
					"children": []any{},
				},
			},
		},
		{
			name: "optional fields all set",
			args: []string{
				"types.TypesService.Optional",
				"--config",
				typesConfigFileName,
				"-d",
				`{` +
					`"optionalString":"test",` +
					`"optionalInt32":42,` +
					`"optionalBool":true,` +
					`"optionalInner":{"id":5},` +
					`"optionalEnum":"COMPLETED"` +
					`}`,
			},
			want: map[string]any{
				"optionalString": "test",
				"optionalInt32":  42.0,
				"optionalBool":   true,
				"optionalInner":  map[string]any{"id": "5"},
				"optionalEnum":   "COMPLETED",
			},
		},
		{
			name: "optional fields none set",
			args: []string{
				"types.TypesService.Optional",
				"--config",
				typesConfigFileName,
				"-d",
				`{}`,
			},
			want: map[string]any{},
		},
		{
			name: "optional fields partially set",
			args: []string{
				"types.TypesService.Optional",
				"--config",
				typesConfigFileName,
				"-d",
				`{"optionalString":"only string","optionalInt32":100}`,
			},
			want: map[string]any{
				"optionalString": "only string",
				"optionalInt32":  100.0,
			},
		},
		{
			name: "repeated fields with values",
			args: []string{
				"types.TypesService.Repeated",
				"--config",
				typesConfigFileName,
				"-d",
				`{` +
					`"strings":["a","b","c"],` +
					`"integers":[1,2,3],` +
					`"doubles":[1.1,2.2,3.3],` +
					`"bools":[true,false,true],` +
					`"statuses":["PENDING","RUNNING","COMPLETED"],` +
					`"innerItems":[{"id":10},{"id":20}],` +
					`"bytesList":["aGVsbG8=","d29ybGQ="]` +
					`}`,
			},
			want: map[string]any{
				"strings":  []any{"a", "b", "c"},
				"integers": []any{"1", "2", "3"},
				"doubles":  []any{1.1, 2.2, 3.3},
				"bools":    []any{true, false, true},
				"statuses": []any{"PENDING", "RUNNING", "COMPLETED"},
				"innerItems": []any{
					map[string]any{"id": "10"},
					map[string]any{"id": "20"},
				},
				"bytesList": []any{"aGVsbG8=", "d29ybGQ="},
			},
		},
		{
			name: "repeated fields empty",
			args: []string{
				"types.TypesService.Repeated",
				"--config",
				typesConfigFileName,
				"-d",
				`{"strings":[],"integers":[],"doubles":[],"bools":[],"statuses":[],"innerItems":[],"bytesList":[]}`,
			},
			want: map[string]any{
				"strings":    []any{},
				"integers":   []any{},
				"doubles":    []any{},
				"bools":      []any{},
				"statuses":   []any{},
				"innerItems": []any{},
				"bytesList":  []any{},
			},
		},
		{
			name: "repeated fields not set",
			args: []string{
				"types.TypesService.Repeated",
				"--config",
				typesConfigFileName,
				"-d",
				`{}`,
			},
			want: map[string]any{
				"strings":    []any{},
				"integers":   []any{},
				"doubles":    []any{},
				"bools":      []any{},
				"statuses":   []any{},
				"innerItems": []any{},
				"bytesList":  []any{},
			},
		},
		{
			name: "repeated fields single values",
			args: []string{
				"types.TypesService.Repeated",
				"--config",
				typesConfigFileName,
				"-d",
				`{` +
					`"strings":["single"],` +
					`"integers":[42],` +
					`"doubles":[3.14],` +
					`"bools":[false],` +
					`"statuses":["FAILED"],` +
					`"innerItems":[{"id":1}],` +
					`"bytesList":["c2luZ2xl"]` +
					`}`,
			},
			want: map[string]any{
				"strings":  []any{"single"},
				"integers": []any{"42"},
				"doubles":  []any{3.14},
				"bools":    []any{false},
				"statuses": []any{"FAILED"},
				"innerItems": []any{
					map[string]any{"id": "1"},
				},
				"bytesList": []any{"c2luZ2xl"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := runCall(fs, nil, tt.args...)
			require.NoErrorf(t, err, "command failed with output: %s", string(b))

			got := map[string]any{}
			require.NoError(t, json.Unmarshal(b, &got))
			require.Equal(t, tt.want, got)
		})
	}
}

func runCall(fs afero.Fs, in io.Reader, args ...string) ([]byte, error) {
	return run(fs, in, nil, append([]string{"call"}, args...)...)
}

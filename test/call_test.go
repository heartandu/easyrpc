package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/heartandu/easyrpc/pkg/app"
)

// TODO: Add streaming call tests, add more protobuf types to tests.
func TestCall(t *testing.T) {
	fileName, cleanup, err := createTempFile("", `{"msg":"file test"}`)
	if err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}
	defer cleanup()

	protoConfigFileName, cleanup, err := createTempFile("*.yaml", `
        address: `+insecureAddress+`
        import_paths:
          - ../internal/testdata
        proto_files:
          - test.proto
    `)
	if err != nil {
		t.Fatalf("failed to create proto config file: %v", err)
	}
	defer cleanup()

	reflectionConfigFileName, cleanup, err := createTempFile("*.yaml", `
        address: `+insecureAddress+`
        reflection: true
    `)
	if err != nil {
		t.Fatalf("failed to create proto config file: %v", err)
	}
	defer cleanup()

	tlsConfigFileName, cleanup, err := createTempFile("*.yaml", `
        address: `+tlsAddress+`
        reflection: true
        tls: true
        cacert: `+cacert+`
        cert: `+cert+`
        cert_key: `+certKey+`
        `)
	if err != nil {
		t.Fatalf("failed to create tls config file: %v", err)
	}
	defer cleanup()

	packageAndServiceConfigFileName, cleanup, err := createTempFile("*.yaml", `
        address: `+insecureAddress+`
        reflection: true
        package: echo
        service: EchoService
    `)
	if err != nil {
		t.Fatalf("failed to create proto config file: %v", err)
	}
	defer cleanup()

	tests := []struct {
		name string
		args []string
		in   io.Reader
		want string
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
				"../internal/testdata",
				"-p",
				"test.proto",
			},
			want: "oops",
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
			want: "hello",
		},
		{
			name: "data from file",
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				insecureAddress,
				"-d",
				"@" + fileName,
				"-r",
			},
			want: "file test",
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
			want: "stdin test",
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
			want: "proto config",
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
			want: "reflection config",
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
			want: "tls",
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
			want: "tls certs",
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
			want: "tls certs config",
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
			want: "package flag",
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
			want: "package and service flags",
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
			want: "package and service flags",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := runCall(tt.in, tt.args...)
			if err != nil {
				t.Fatalf("command failed: output = %v, err = %v", string(b), err)
			}

			result := map[string]string{}

			d := json.NewDecoder(bytes.NewReader(b))
			if err := d.Decode(&result); err != nil {
				t.Fatalf("failed to decode output: %v", err)
			}

			if result["msg"] != tt.want {
				t.Fatalf("unexpected response: got = %v, want = %v", result["msg"], tt.want)
			}
		})
	}
}

func createTempFile(pattern, contents string) (string, func(), error) { //nolint:gocritic // no need to name returns
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", func() {}, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer f.Close()

	if _, err = f.WriteString(contents); err != nil {
		return "", func() {}, fmt.Errorf("failed to write contents: %w", err)
	}

	return f.Name(), func() {
		if err := os.Remove(f.Name()); err != nil {
			log.Printf("failed to remove input file: %v", err)
		}
	}, nil
}

func runCall(in io.Reader, args ...string) ([]byte, error) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = append([]string{
		"easyrpc",
		"call",
	}, args...)

	b := bytes.NewBuffer(nil)

	a := app.NewApp()
	a.SetOutput(b)

	if in != nil {
		a.SetInput(in)
	}

	err := a.Run()

	return b.Bytes(), err
}

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
        server:
          address: `+address+`
        proto:
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
        server:
          address: `+address+`
          reflection: true
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
				"-a",
				address,
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
				"-a",
				address,
				"-d",
				`{"msg":"hello"}`,
				"-r",
			},
			want: "hello",
		},
		{
			name: "data from file",
			args: []string{
				"-a",
				address,
				"-d",
				"@" + fileName,
				"-r",
			},
			want: "file test",
		},
		{
			name: "data from stdin",
			args: []string{
				"-a",
				address,
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
				"--config",
				reflectionConfigFileName,
				"-d",
				`{"msg":"reflection config"}`,
			},
			want: "reflection config",
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
		"echo.EchoService.Echo",
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

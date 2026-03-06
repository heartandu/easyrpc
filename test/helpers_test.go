package test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/heartandu/easyrpc/internal/app"
)

func homeDir(t *testing.T) string {
	t.Helper()

	home, err := os.UserHomeDir()
	require.NoError(t, err, "failed to get home directory")

	return home
}

func createTempFile(fs afero.Fs, name, contents string) (string, error) {
	file, err := fs.Create(filepath.Join(afero.GetTempDir(fs, ""), name))
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer file.Close()

	if _, err = file.WriteString(contents); err != nil {
		return "", fmt.Errorf("failed to write contents: %w", err)
	}

	return file.Name(), nil
}

func createFileAtPath(fs afero.Fs, path, contents string) error {
	dir := filepath.Dir(path)
	if err := fs.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := fs.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err = file.WriteString(contents); err != nil {
		return fmt.Errorf("failed to write contents: %w", err)
	}

	return nil
}

func run(fs afero.Fs, input io.Reader, env map[string]string, args ...string) ([]byte, error) {
	oldArgs := os.Args

	defer func() { os.Args = oldArgs }()

	os.Args = append([]string{"easyrpc"}, args...)

	oldEnv := make(map[string]string)
	for k, v := range env {
		oldEnv[k] = os.Getenv(k)
		if err := os.Setenv(k, v); err != nil {
			return nil, fmt.Errorf("failed to set env var %s: %w", k, err)
		}
	}

	defer func() {
		for k, v := range oldEnv {
			if err := os.Setenv(k, v); err != nil {
				fmt.Fprintf(os.Stderr, "failed to restore env var %s: %v\n", k, err)
			}
		}
	}()

	buf := bytes.NewBuffer(nil)

	a := app.NewApp("0.0.0-SNAPSHOT")
	a.SetOutput(buf)
	a.SetFs(fs)

	if input != nil {
		a.SetInput(input)
	}

	err := a.Run()

	return buf.Bytes(), err
}

func address(socket string) string {
	return "localhost" + socket
}

package test

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/spf13/afero"

	"github.com/heartandu/easyrpc/pkg/app"
)

func createTempFile(fs afero.Fs, name, contents string) (string, error) {
	file, err := fs.Create(name)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer file.Close()

	if _, err = file.WriteString(contents); err != nil {
		return "", fmt.Errorf("failed to write contents: %w", err)
	}

	return file.Name(), nil
}

func run(fs afero.Fs, input io.Reader, args ...string) ([]byte, error) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = append([]string{"easyrpc"}, args...)

	buf := bytes.NewBuffer(nil)

	a := app.NewApp()
	a.SetOutput(buf)
	a.SetFs(fs)

	if input != nil {
		a.SetInput(input)
	}

	err := a.Run()

	return buf.Bytes(), err
}

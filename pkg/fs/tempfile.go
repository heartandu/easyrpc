package fs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// CreateTempFile makes a temporary directory named after the command name,
// and puts a new temporary file in that directory for reading and writing.
func CreateTempFile(fs afero.Fs, pattern string) (afero.File, error) {
	dirname := filepath.Base(os.Args[0])
	tempDir := afero.GetTempDir(fs, dirname)

	f, err := afero.TempFile(fs, tempDir, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	return f, nil
}

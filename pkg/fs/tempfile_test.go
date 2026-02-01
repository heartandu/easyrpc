package fs_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/heartandu/easyrpc/pkg/fs"
)

//nolint:paralleltest // test sets global os.Args[0] which can't be done in parallel
func TestCreateTempFile_SuccessfulCreation(t *testing.T) {
	memFs := afero.NewMemMapFs()

	tests := []struct {
		name       string
		binaryName string
	}{
		{
			name: "default full path",
		},
		{
			name: "custom full path",
			binaryName: func() string {
				p, err := filepath.Abs("easyrpc")
				require.NoError(t, err)

				return p
			}(),
		},
		{
			name:       "short name",
			binaryName: "easyrpc",
		},
		{
			name:       "local path",
			binaryName: filepath.Join("bin", "easyrpc"),
		},
	}

	//nolint:paralleltest // test sets global os.Args[0] which can't be done in parallel
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.binaryName != "" {
				oldBinaryName := os.Args[0]
				os.Args[0] = tt.binaryName

				t.Cleanup(func() { os.Args[0] = oldBinaryName })
			}

			f, err := fs.CreateTempFile(memFs, "test-*.txt")
			require.NoError(t, err)
			require.NotNil(t, f)

			t.Cleanup(func() { f.Close() })

			name := f.Name()
			require.Contains(t, filepath.Base(name), "test-")
			require.Contains(t, filepath.Base(name), ".txt")

			tempdir := afero.GetTempDir(memFs, "")
			tempdir = filepath.Join(tempdir, filepath.Base(os.Args[0]))
			require.Equal(t, tempdir, filepath.Dir(name))

			exists, err := afero.Exists(memFs, name)
			require.NoError(t, err)
			require.True(t, exists)
		})
	}
}

//nolint:paralleltest // test uses global os.Args[0] which is being set in other tests
func TestCreateTempFile_UniqueNames(t *testing.T) {
	memFs := afero.NewMemMapFs()

	files := make(map[string]afero.File)

	for range 3 {
		f, err := fs.CreateTempFile(memFs, "prefix-*.log")
		require.NoError(t, err)
		require.NotNil(t, f)

		t.Cleanup(func() { f.Close() })

		files[f.Name()] = f
	}

	require.Len(t, files, 3, "all files should have unique names")

	for _, f := range files {
		exists, err := afero.Exists(memFs, f.Name())
		require.NoError(t, err)
		require.True(t, exists)
	}
}

//nolint:paralleltest // test uses global os.Args[0] which is being set in other tests
func TestCreateTempFile_VariousPatterns(t *testing.T) {
	memFs := afero.NewMemMapFs()

	tests := []struct {
		name    string
		pattern string
		prefix  string
		suffix  string
	}{
		{
			name: "empty pattern",
		},
		{
			name:    "suffix only",
			pattern: "*.tmp",
			suffix:  ".tmp",
		},
		{
			name:    "prefix only",
			pattern: "prefix-*",
			prefix:  "prefix-",
		},
		{
			name:    "suffix with extension",
			pattern: "*-suffix.log",
			suffix:  "-suffix.log",
		},
		{
			name:    "prefix and suffix",
			pattern: "prefix-*-suffix.txt",
			prefix:  "prefix-",
			suffix:  "-suffix.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := fs.CreateTempFile(memFs, tt.pattern)
			require.NoError(t, err)
			require.NotNil(t, f)

			t.Cleanup(func() { f.Close() })

			basename := filepath.Base(f.Name())
			if tt.prefix != "" {
				require.Contains(t, basename, tt.prefix)
			}

			if tt.suffix != "" {
				require.Contains(t, basename, tt.suffix)
			}

			exists, err := afero.Exists(memFs, f.Name())
			require.NoError(t, err)
			require.True(t, exists)
		})
	}
}

//nolint:paralleltest // test uses global os.Args[0] which is being set in other tests
func TestCreateTempFile_FileOperations(t *testing.T) {
	memFs := afero.NewMemMapFs()

	f, err := fs.CreateTempFile(memFs, "data-*.txt")
	require.NoError(t, err)
	require.NotNil(t, f)

	t.Cleanup(func() { f.Close() })

	testData := []byte("test content for temp file")
	n, err := f.Write(testData)
	require.NoError(t, err)
	require.Equal(t, len(testData), n)

	_, err = f.Seek(0, 0)
	require.NoError(t, err)

	readData, err := io.ReadAll(f)
	require.NoError(t, err)
	require.Equal(t, testData, readData)

	info, err := f.Stat()
	require.NoError(t, err)
	require.Equal(t, int64(len(testData)), info.Size())
}

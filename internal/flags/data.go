package flags

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	fsutil "github.com/heartandu/easyrpc/pkg/fs"
)

const (
	dataFlag = "data"
	fileFlag = "file"
)

var ErrDataAndFileFlagDisallowed = errors.New("only data or file flag is allowed to be set")

// RegisterDataAndFileFlags registers the data and the file flags with the provided command.
// The flags allow the user to specify request data by providing raw data string, read the
// data from stdin or from a file.
func RegisterDataAndFileFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(dataFlag, "d", "", "request data in json format (specify \"-\" to read stdin)")
	cmd.Flags().StringP(fileFlag, "f", "", "file name with request data in json format")
}

// HandleDataOrFileFlag returns an io.ReadCloser for the data specified in the data or the file flag.
// If the data flag is "-", the data will be read from stdin.
// If the file flag is specified, open the file.
// If no flag is set, returns the provided data string.
// If the data flag and the file flag are both set, return an error.
func HandleDataOrFileFlag(cmd *cobra.Command, fs afero.Fs) (io.ReadCloser, error) {
	data, err := cmd.Flags().GetString(dataFlag)
	if err != nil {
		return nil, fmt.Errorf("failed to get data flag: %w", err)
	}

	filename, err := cmd.Flags().GetString(fileFlag)
	if err != nil {
		return nil, fmt.Errorf("failed to get file flag: %w", err)
	}

	if data != "" && filename != "" {
		return nil, ErrDataAndFileFlagDisallowed
	}

	if data == "-" {
		return io.NopCloser(cmd.InOrStdin()), nil
	}

	if filename != "" {
		path, err := fsutil.ExpandHome(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to expand home dir: %w", err)
		}

		file, err := fs.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}

		return file, nil
	}

	return io.NopCloser(bytes.NewReader([]byte(data))), nil
}

package flags

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// RegisterDataFlag registers the data flag with the provided command.
// The flag allows the user to specify request data.
func RegisterDataFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("data", "d", "", "request data in json format")
}

// HandleDataFlag returns an io.ReadCloser for the data specified in the flag.
// If the data flag is "-", the data will be read from stdin.
// If the data flag starts with "@", it opens the file specified in the flag.
// Otherwise, it returns the provided data string.
func HandleDataFlag(cmd *cobra.Command, fs afero.Fs) (io.ReadCloser, error) {
	data, err := cmd.Flags().GetString("data")
	if err != nil {
		return nil, fmt.Errorf("failed to get data flag: %w", err)
	}

	if data == "-" {
		return io.NopCloser(cmd.InOrStdin()), nil
	}

	if strings.HasPrefix(data, "@") {
		file, err := fs.Open(data[1:])
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}

		return file, nil
	}

	return io.NopCloser(bytes.NewReader([]byte(data))), nil
}

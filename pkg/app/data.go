package app

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func registerDataFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("data", "d", "", "request data in json format")
}

func handleDataFlag(fs afero.Fs, cmd *cobra.Command) (io.ReadCloser, error) {
	data, err := cmd.Flags().GetString("data")
	if err != nil {
		return nil, fmt.Errorf("failed to get data flag: %w", err)
	}

	if strings.HasPrefix(data, "@") {
		if len(data) == 1 {
			return io.NopCloser(cmd.InOrStdin()), nil
		}

		file, err := fs.Open(data[1:])
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}

		return file, nil
	}

	return io.NopCloser(bytes.NewReader([]byte(data))), nil
}

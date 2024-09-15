package flags

import (
	"fmt"
	"io"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// RegisterOutputFlag registers the output flag for a given command.
// The flag allows the user to specify an output file to write to.
func RegisterOutputFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("output", "o", "", "output file to write to")
}

// HandleOutputFlag returns an io.WriteCloser that represents the output destination.
func HandleOutputFlag(cmd *cobra.Command, fs afero.Fs) (io.WriteCloser, error) {
	out := nopWriteCloser(cmd.OutOrStdout())

	outFile, err := cmd.Flags().GetString("output")
	if err != nil {
		return nil, fmt.Errorf("failed to get output flag: %w", err)
	}

	if outFile != "" {
		out, err = fs.Create(outFile)
		if err != nil {
			return nil, fmt.Errorf("failed to open output file: %w", err)
		}
	}

	return out, nil
}

// nopWriteCloser returns a no-op closer that does not actually close the underlying writer.
func nopWriteCloser(w io.Writer) io.WriteCloser {
	return &nopCloser{w: w}
}

// nopCloser is a no-op closer wrapper implementation of io.WriteCloser.
type nopCloser struct {
	w io.Writer
}

// Write writes the provided bytes to the underlying writer.
func (w *nopCloser) Write(p []byte) (int, error) {
	return w.w.Write(p) //nolint:wrapcheck // This is a simple decorator.
}

// Close does nothing and always returns nil, as this is a no-op closer.
func (*nopCloser) Close() error {
	return nil
}

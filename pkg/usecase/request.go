package usecase

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/afero"

	"github.com/heartandu/easyrpc/pkg/descriptor"
	"github.com/heartandu/easyrpc/pkg/editor"
	"github.com/heartandu/easyrpc/pkg/format"
)

// Request represents a use case for populating a request message to stdout or a file.
type Request struct {
	out    io.Writer
	editor editor.Editor
	fs     afero.Fs
	ds     descriptor.Source
	mf     format.MessageFormatter
}

// NewRequest returns a new instance of Request.
func NewRequest(
	out io.Writer,
	e editor.Editor,
	fs afero.Fs,
	ds descriptor.Source,
	mf format.MessageFormatter,
) *Request {
	return &Request{
		out:    out,
		editor: e,
		fs:     fs,
		ds:     ds,
		mf:     mf,
	}
}

// Prepare formats a request message for the specified method,
// and optionally allows editing it before writing it to an output.
func (r *Request) Prepare(method string) error {
	m, err := r.ds.FindMethod(method)
	if err != nil {
		return fmt.Errorf("failed to find method: %w", err)
	}

	msg, err := r.mf.Format(m.RequestMessage())
	if err != nil {
		return fmt.Errorf("failed to format message: %w", err)
	}

	if r.editor != nil {
		msg, err = r.editor.Run(msg)
		if err != nil {
			return fmt.Errorf("failed to edit the message: %w", err)
		}
	}

	fmt.Fprintf(r.out, "%v\n", strings.TrimSpace(msg))

	return nil
}

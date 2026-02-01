package usecase

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/heartandu/protoreflect-jsonschema/protoschema"
	"github.com/spf13/afero"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/heartandu/easyrpc/pkg/descriptor"
	"github.com/heartandu/easyrpc/pkg/editor"
	"github.com/heartandu/easyrpc/pkg/format"
	fsutils "github.com/heartandu/easyrpc/pkg/fs"
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
// Editing a message is in JSON format only, but the actual output format
// will depend on a message formatter.
func (r *Request) Prepare(ctx context.Context, method string) error {
	m, err := r.ds.FindMethod(method)
	if err != nil {
		return fmt.Errorf("failed to find method: %w", err)
	}

	msg := m.RequestMessage()
	if err = r.acceptMessage(ctx, msg); err != nil {
		return fmt.Errorf("failed to accept message: %w", err)
	}

	formatted, err := r.mf.Format(msg)
	if err != nil {
		return fmt.Errorf("failed to format message: %w", err)
	}

	fmt.Fprintf(r.out, "%s\n", strings.TrimSpace(formatted))

	return nil
}

func (r *Request) acceptMessage(ctx context.Context, msg proto.Message) error {
	if r.editor == nil {
		return nil
	}

	schemaFileName, err := r.writeTempSchema(msg)
	if err != nil {
		return fmt.Errorf("failed to write temporary jsonschema: %w", err)
	}

	defer func() {
		if err = r.fs.Remove(schemaFileName); err != nil {
			// TODO: pass a logger instance instead of calling the global one.
			log.Printf("couldn't remove temporary file: %v", err)
		}
	}()

	messageToAccept := fmt.Sprintf("{\n  \"$schema\": %q\n}", "file://"+schemaFileName)

	accepted, err := r.editor.Run(ctx, messageToAccept)
	if err != nil {
		return fmt.Errorf("failed to edit the message: %w", err)
	}
	defer accepted.Close()

	mp := format.JSONMessageParser(accepted, protojson.UnmarshalOptions{DiscardUnknown: true})
	if err := mp.Next(msg); err != nil {
		return fmt.Errorf("failed to parse accepted message: %w", err)
	}

	return nil
}

func (r *Request) writeTempSchema(msg proto.Message) (string, error) {
	file, err := fsutils.CreateTempFile(r.fs, "*.schema.json")
	if err != nil {
		return "", fmt.Errorf("failed to create temp schema file: %w", err)
	}
	defer file.Close()

	if err := protoschema.NewEncoder(file).Encode(msg.ProtoReflect()); err != nil {
		return "", fmt.Errorf("failed to encode jsonschema: %w", err)
	}

	return file.Name(), nil
}

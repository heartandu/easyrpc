package usecase

import (
	"context"
	"encoding/json"
	"errors"
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
	mp     format.MessageParser
	mf     format.MessageFormatter
}

// NewRequest returns a new instance of Request.
func NewRequest(
	out io.Writer,
	e editor.Editor,
	fs afero.Fs,
	ds descriptor.Source,
	mp format.MessageParser,
	mf format.MessageFormatter,
) *Request {
	return &Request{
		out:    out,
		editor: e,
		fs:     fs,
		ds:     ds,
		mp:     mp,
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

	schemaFileName, cleanup, err := r.writeTempSchema(msg)
	if err != nil {
		return fmt.Errorf("failed to write temporary jsonschema file: %w", err)
	}
	defer cleanup()

	readMsgsCount := 0

readLoop:
	for {
		err = r.mp.Next(msg)
		switch {
		case errors.Is(err, io.EOF):
			// Process a message at least once
			if readMsgsCount > 0 {
				break readLoop
			}
		case err != nil:
			return fmt.Errorf("failed to parse a message: %w", err)
		default:
		}

		if err = r.acceptMessage(ctx, msg, schemaFileName); err != nil {
			return fmt.Errorf("failed to accept message: %w", err)
		}

		formatted, err := r.mf.Format(msg)
		if err != nil {
			return fmt.Errorf("failed to format message: %w", err)
		}

		fmt.Fprintf(r.out, "%s\n", strings.TrimSpace(formatted))

		readMsgsCount++
	}

	return nil
}

func (r *Request) writeTempSchema(msg proto.Message) (filename string, cleanup func(), err error) {
	if r.editor == nil {
		return "", func() {}, nil
	}

	file, err := fsutils.CreateTempFile(r.fs, "*.schema.json")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp schema file: %w", err)
	}
	defer file.Close()

	if err = protoschema.NewEncoder(file).Encode(msg.ProtoReflect()); err != nil {
		return "", nil, fmt.Errorf("failed to encode jsonschema: %w", err)
	}

	filename = file.Name()

	return filename, func() {
		if err := r.fs.Remove(filename); err != nil {
			// TODO: pass a logger instance instead of calling the global one.
			log.Printf("couldn't remove temporary file: %v", err)
		}
	}, nil
}

func (r *Request) acceptMessage(ctx context.Context, msg proto.Message, schema string) error {
	if r.editor == nil {
		return nil
	}

	messageToAccept, err := r.injectSchemaAndMarshal(msg, schema)
	if err != nil {
		return fmt.Errorf("failed to marshame message to accept: %w", err)
	}

	accepted, err := r.editor.Run(ctx, messageToAccept)
	if err != nil {
		return fmt.Errorf("failed to edit the message: %w", err)
	}
	defer accepted.Close()

	mp := format.JSONMessageParser(accepted, protojson.UnmarshalOptions{DiscardUnknown: true})
	if err := mp.Next(msg); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to parse accepted message: %w", err)
	}

	return nil
}

// NOTE: This is horribly inefficient, but I have no idea how to inject $schema to a message in one pass.
func (*Request) injectSchemaAndMarshal(msg proto.Message, schema string) (string, error) {
	messageBytes, err := protojson.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("failed to preformat message: %w", err)
	}

	msgMap := map[string]any{}
	if err = json.Unmarshal(messageBytes, &msgMap); err != nil {
		return "", fmt.Errorf("failed to unmarshal preformatted message: %w", err)
	}

	msgMap["$schema"] = "file://" + schema

	resultingMessage, err := json.MarshalIndent(msgMap, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshame message to accept: %w", err)
	}

	return string(resultingMessage), nil
}

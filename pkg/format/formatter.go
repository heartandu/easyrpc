package format

import (
	"encoding/json"
	"fmt"

	"github.com/heartandu/protoreflect-jsonschema/protoschema"
	"github.com/spf13/afero"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// MessageFormatter is an interface that defines a method for formatting a protobuf message into a string.
type MessageFormatter interface {
	Format(msg proto.Message) (string, error)
}

// JSONMessageFormatter creates a new MessageFormatter that formats messages as JSON using the provided MarshalOptions.
func JSONMessageFormatter(out protojson.MarshalOptions) MessageFormatter {
	return &jsonMessageFormatter{
		out: out,
	}
}

type jsonMessageFormatter struct {
	out protojson.MarshalOptions
}

// Format formats the given protobuf message as a JSON string using the MarshalOptions provided during creation.
func (f *jsonMessageFormatter) Format(msg proto.Message) (string, error) {
	return f.out.Format(msg), nil
}

// JSONSchemaMessageFormatter creates a new MessageFormatter that produces a temporary jsonschema file,
// and makes a JSON request file with that schema.
func JSONSchemaMessageFormatter(fs afero.Fs) MessageFormatter {
	return &jsonschemaMessageFormatter{fs: fs}
}

type jsonschemaMessageFormatter struct {
	fs afero.Fs
}

func (f *jsonschemaMessageFormatter) Format(msg proto.Message) (string, error) {
	schemaFileName, err := f.writeTempSchema(msg)
	if err != nil {
		return "", fmt.Errorf("failed to write temp schema: %w", err)
	}

	out := struct {
		Schema string `json:"$schema"`
	}{Schema: schemaFileName}

	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to encode message: %w", err)
	}

	return string(b), nil
}

func (f *jsonschemaMessageFormatter) writeTempSchema(msg proto.Message) (string, error) {
	tempDir := afero.GetTempDir(f.fs, "easyrpc")

	file, err := afero.TempFile(f.fs, tempDir, "*.schema.json")
	if err != nil {
		return "", fmt.Errorf("failed to create temp schema file: %w", err)
	}
	defer file.Close()

	if err := protoschema.NewEncoder(file).Encode(msg.ProtoReflect()); err != nil {
		return "", fmt.Errorf("failed to encode jsonschema: %w", err)
	}

	return file.Name(), nil
}

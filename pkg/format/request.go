package format

import (
	"encoding/json"
	"fmt"
	"io"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// RequestParser is an interface for parsing requests.
type RequestParser interface {
	Next(msg proto.Message) error
}

// JSONRequestParser creates a new RequestParser for JSON input.
func JSONRequestParser(input io.ReadCloser, unmarshalOpts protojson.UnmarshalOptions) RequestParser {
	return &jsonRequestParser{
		decoder: json.NewDecoder(input),
		out:     unmarshalOpts,
	}
}

type jsonRequestParser struct {
	decoder *json.Decoder
	out     protojson.UnmarshalOptions
}

// Next reads and unmarshals JSON input into a proto.Message.
func (p *jsonRequestParser) Next(msg proto.Message) error {
	var raw json.RawMessage
	if err := p.decoder.Decode(&raw); err != nil {
		return fmt.Errorf("failed to read raw input: %w", err)
	}

	if err := p.out.Unmarshal(raw, msg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return nil
}

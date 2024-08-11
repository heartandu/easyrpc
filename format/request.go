package format

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// RequestParser is an interface for parsing requests.
type RequestParser interface {
	Parse(msg proto.Message) error
}

// JSONRequestParser creates a new RequestParser for JSON input.
func JSONRequestParser(input io.Reader, unmarshalOpts protojson.UnmarshalOptions) RequestParser {
	return &jsonRequestParser{
		in:  json.NewDecoder(input),
		out: unmarshalOpts,
	}
}

type jsonRequestParser struct {
	in  *json.Decoder
	out protojson.UnmarshalOptions
}

// Parse reads and unmarshals JSON input into a proto.Message.
func (p *jsonRequestParser) Parse(msg proto.Message) error {
	var raw json.RawMessage
	if err := p.in.Decode(&raw); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}

		return fmt.Errorf("failed to read raw input: %w", err)
	}

	if err := p.out.Unmarshal(raw, msg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return nil
}

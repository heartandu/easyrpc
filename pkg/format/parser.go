package format

import (
	"encoding/json"
	"fmt"
	"io"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// MessageParser is an interface for parsing requests.
type MessageParser interface {
	Next(msg proto.Message) error
}

// JSONMessageParser creates a new MessageParser for JSON input.
func JSONMessageParser(input io.Reader, unmarshalOpts protojson.UnmarshalOptions) MessageParser {
	return &jsonMessageParser{
		decoder: json.NewDecoder(input),
		out:     unmarshalOpts,
	}
}

type jsonMessageParser struct {
	decoder *json.Decoder
	out     protojson.UnmarshalOptions
}

// Next reads and unmarshals JSON input into a proto.Message.
func (p *jsonMessageParser) Next(msg proto.Message) error {
	var raw json.RawMessage
	if err := p.decoder.Decode(&raw); err != nil {
		return fmt.Errorf("failed to read raw input: %w", err)
	}

	if err := p.out.Unmarshal(raw, msg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return nil
}

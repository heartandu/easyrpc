package format

import (
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

package format

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ResponseFormatter is an interface that defines a method for formatting a protobuf message into a string.
type ResponseFormatter interface {
	Format(msg proto.Message) (string, error)
}

// JSONResponseFormatter creates a new ResponseFormatter that formats messages as JSON using
// the provided MarshalOptions.
func JSONResponseFormatter(out protojson.MarshalOptions) ResponseFormatter {
	return &jsonResponseFormatter{
		out: out,
	}
}

type jsonResponseFormatter struct {
	out protojson.MarshalOptions
}

// Format formats the given protobuf message as a JSON string using the MarshalOptions provided during creation.
func (f *jsonResponseFormatter) Format(msg proto.Message) (string, error) {
	return f.out.Format(msg), nil
}

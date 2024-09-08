package format_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/heartandu/easyrpc/internal/testdata"
	"github.com/heartandu/easyrpc/pkg/format"
)

func TestJSONRequestParser_Parse(t *testing.T) {
	t.Parallel()

	testErr := errors.New("oh no")

	tests := []struct {
		name    string
		input   io.Reader
		want    *testdata.EchoRequest
		wantErr error
	}{
		{
			name:    "success",
			input:   strings.NewReader(`{"msg":"hi"}`),
			want:    &testdata.EchoRequest{Msg: "hi"},
			wantErr: nil,
		},
		{
			name:    "empty message",
			input:   strings.NewReader(""),
			want:    &testdata.EchoRequest{},
			wantErr: nil,
		},
		{
			name: "reader error",
			input: funcReader(func(p []byte) (int, error) {
				return 0, testErr
			}),
			want:    &testdata.EchoRequest{},
			wantErr: testErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parser := format.JSONRequestParser(tt.input, protojson.UnmarshalOptions{})

			got := &testdata.EchoRequest{}

			err := parser.Parse(got)
			require.ErrorIs(t, err, tt.wantErr)

			if !proto.Equal(got, tt.want) {
				t.Errorf("Parse() got = %v, want = %v", got, tt.want)
			}
		})
	}
}

type funcReader func(p []byte) (int, error)

func (f funcReader) Read(p []byte) (int, error) {
	return f(p)
}
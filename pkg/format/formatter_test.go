package format_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/heartandu/easyrpc/internal/testdata"
	"github.com/heartandu/easyrpc/pkg/format"
)

func TestJSONMessageFormatter_Format(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		out  protojson.MarshalOptions
		msg  *testdata.EchoResponse
		want string
	}{
		{
			name: "default",
			out:  protojson.MarshalOptions{},
			msg:  &testdata.EchoResponse{Msg: "hi"},
			want: `{"msg":"hi"}`,
		},
		{
			name: "multiline",
			out:  protojson.MarshalOptions{Multiline: true},
			msg:  &testdata.EchoResponse{Msg: "hi"},
			want: `{
  "msg": "hi"
}`,
		},
		{
			name: "custom indent",
			out:  protojson.MarshalOptions{Indent: "    "},
			msg:  &testdata.EchoResponse{Msg: "hi"},
			want: `{
    "msg": "hi"
}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			formatter := format.JSONMessageFormatter(tt.out)

			got, err := formatter.Format(tt.msg)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

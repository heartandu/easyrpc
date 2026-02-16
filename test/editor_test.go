package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestEditorFlag(t *testing.T) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		t.Skip("Skipping test: no TTY available")
	}

	tty.Close()

	fs := afero.NewOsFs()

	mockEditorFile, err := createTempFile(fs, "mock_editor.sh", `
if ! grep -q '$schema' "$1"; then
	echo "Error: \$schema field not found in input" >&2
	exit 1
fi
echo '{"msg": "edited message"}' > "$1"`)
	require.NoError(t, err)
	t.Cleanup(func() { fs.Remove(mockEditorFile) })

	mockEditorLongFlagFile, err := createTempFile(fs, "mock_editor_long.sh", `
if ! grep -q '$schema' "$1"; then
	echo "Error: \$schema field not found in input" >&2
	exit 1
fi
echo '{"msg": "long flag edited"}' > "$1"`)
	require.NoError(t, err)
	t.Cleanup(func() { fs.Remove(mockEditorLongFlagFile) })

	mockEditorStreamingFile, err := createTempFile(fs, "mock_editor_streaming.sh", `
if ! grep -q '$schema' "$1"; then
	echo "Error: \$schema field not found in input" >&2
	exit 1
fi
msg_value=$(grep -o '"msg": *"[^"]*"' "$1" | sed 's/.*"msg": *"\([^"]*\)".*/\1/')
echo "{\"msg\": \"${msg_value}_edited\"}" > "$1"`)
	require.NoError(t, err)
	t.Cleanup(func() { fs.Remove(mockEditorStreamingFile) })

	tests := []struct {
		name    string
		command string
		editor  string
		args    []string
		want    []map[string]any
	}{
		{
			name:    "call with short editor flag",
			command: "call",
			editor:  "sh " + mockEditorFile,
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				address(insecureSocket),
				"-r",
				"-e",
			},
			want: []map[string]any{{"msg": "edited message"}},
		},
		{
			name:    "call with long editor flag",
			command: "call",
			editor:  "sh " + mockEditorLongFlagFile,
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				address(insecureSocket),
				"-r",
				"--edit",
			},
			want: []map[string]any{{"msg": "long flag edited"}},
		},
		{
			name:    "call client streaming with editor edits all messages",
			command: "call",
			editor:  "sh " + mockEditorStreamingFile,
			args: []string{
				"echo.EchoService.ClientStream",
				"-a",
				address(insecureSocket),
				"-r",
				"-e",
				"-d",
				`{"msg":"1"}{"msg":"2"}{"msg":"3"}`,
			},
			want: []map[string]any{{"msgs": []any{"1_edited", "2_edited", "3_edited"}}},
		},
		{
			name:    "call bidi streaming with editor edits all messages",
			command: "call",
			editor:  "sh " + mockEditorStreamingFile,
			args: []string{
				"echo.EchoService.BidiStream",
				"-a",
				address(insecureSocket),
				"-r",
				"-e",
				"-d",
				`{"msg":"a"}{"msg":"b"}`,
			},
			want: []map[string]any{{"msg": "a_edited"}, {"msg": "b_edited"}},
		},
		{
			name:    "call client streaming with single message",
			command: "call",
			editor:  "sh " + mockEditorStreamingFile,
			args: []string{
				"echo.EchoService.ClientStream",
				"-a",
				address(insecureSocket),
				"-r",
				"-e",
				"-d",
				`{"msg":"single"}`,
			},
			want: []map[string]any{{"msgs": []any{"single_edited"}}},
		},
		{
			name:    "request with short editor flag",
			command: "request",
			editor:  "sh " + mockEditorFile,
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				address(insecureSocket),
				"-r",
				"-e",
			},
			want: []map[string]any{{"msg": "edited message"}},
		},
		{
			name:    "request with long editor flag",
			command: "request",
			editor:  "sh " + mockEditorLongFlagFile,
			args: []string{
				"echo.EchoService.Echo",
				"-a",
				address(insecureSocket),
				"-r",
				"--edit",
			},
			want: []map[string]any{{"msg": "long flag edited"}},
		},
		{
			name:    "request client streaming with editor edits all messages",
			command: "request",
			editor:  "sh " + mockEditorStreamingFile,
			args: []string{
				"echo.EchoService.ClientStream",
				"-a",
				address(insecureSocket),
				"-r",
				"-e",
				"-d",
				`{"msg":"1"}{"msg":"2"}{"msg":"3"}`,
			},
			want: []map[string]any{{"msg": "1_edited"}, {"msg": "2_edited"}, {"msg": "3_edited"}},
		},
		{
			name:    "request bidi streaming with editor edits all messages",
			command: "request",
			editor:  "sh " + mockEditorStreamingFile,
			args: []string{
				"echo.EchoService.BidiStream",
				"-a",
				address(insecureSocket),
				"-r",
				"-e",
				"-d",
				`{"msg":"a"}{"msg":"b"}`,
			},
			want: []map[string]any{{"msg": "a_edited"}, {"msg": "b_edited"}},
		},
		{
			name:    "request client streaming with single message",
			command: "request",
			editor:  "sh " + mockEditorStreamingFile,
			args: []string{
				"echo.EchoService.ClientStream",
				"-a",
				address(insecureSocket),
				"-r",
				"-e",
				"-d",
				`{"msg":"single"}`,
			},
			want: []map[string]any{{"msg": "single_edited"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := runWithEditor(fs, tt.command, tt.editor, tt.args...)
			require.NoErrorf(t, err, "command failed with output: %s", string(b))

			got := []map[string]any{}
			d := json.NewDecoder(bytes.NewReader(b))

			for {
				v := map[string]any{}
				if err := d.Decode(&v); err != nil {
					if errors.Is(err, io.EOF) {
						break
					}

					t.Fatalf("failed to decode output: %v", err)
				}

				got = append(got, v)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func runWithEditor(fs afero.Fs, command, editor string, args ...string) ([]byte, error) {
	env := map[string]string{
		"EDITOR": editor,
	}

	return run(fs, nil, env, append([]string{command}, args...)...)
}

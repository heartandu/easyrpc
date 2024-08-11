package fs_test

import (
	"os/user"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/heartandu/easyrpc/pkg/fs"
)

func TestExpandHome(t *testing.T) {
	t.Parallel()

	u, err := user.Current()
	require.NoError(t, err)

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "absolute path",
			path: "/usr/local/bin",
			want: "/usr/local/bin",
		},
		{
			name: "relative path",
			path: "some/path",
			want: "some/path",
		},
		{
			name: "relative path starting with a dot",
			path: "./some/path",
			want: "./some/path",
		},
		{
			name: "relative path starting from a previous dir",
			path: "../some/path",
			want: "../some/path",
		},
		{
			name: "path starting with tilde",
			path: "~/some/path",
			want: u.HomeDir + "/some/path",
		},
		{
			name: "single tilde",
			path: "~",
			want: u.HomeDir,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := fs.ExpandHome(tt.path)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

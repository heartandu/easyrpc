package fs

import (
	"fmt"
	"os/user"
	"path/filepath"
	"strings"
)

// ExpandHome expands a path that starts with a tilde (~) to the user's home directory.
// If the path does not start with a tilde, it returns the original path.
func ExpandHome(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	curUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to fetch current user: %w", err)
	}

	if len(path) == 1 {
		return curUser.HomeDir, nil
	}

	return filepath.Join(curUser.HomeDir, path[2:]), nil
}

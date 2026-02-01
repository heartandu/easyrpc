package editor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/afero"

	fsutils "github.com/heartandu/easyrpc/pkg/fs"
)

// ErrNoCmd is an error returned when no command is provided to run.
var ErrNoCmd = errors.New("no command to run")

// Editor is an interface that allows message editing in an arbitrary text editor.
type Editor interface {
	Run(ctx context.Context, input string) (io.ReadCloser, error)
}

// NewCmdEditor creates a new instance of cmdEditor which runs an editor specified in cmd.
func NewCmdEditor(fs afero.Fs, cmd string) Editor {
	return &cmdEditor{
		cmd: cmd,
		fs:  fs,
	}
}

type cmdEditor struct {
	cmd string
	fs  afero.Fs
}

// Run executes the editor command with the provided message and returns the resulting output.
func (e *cmdEditor) Run(ctx context.Context, input string) (io.ReadCloser, error) {
	fileName, err := e.writeInputToFile(input)
	defer e.cleanUp(fileName)

	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	cmdArgs := strings.Split(e.cmd, " ")
	if len(cmdArgs) < 1 {
		return nil, ErrNoCmd
	}

	cmdArgs = append(cmdArgs, fileName)

	// TODO: make this cross-platform
	// Open new terminal stdin and stdout
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open separate tty: %w", err)
	}
	defer tty.Close()

	//nolint:gosec // This should be fine if ran on a user machine.
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdin = tty
	cmd.Stdout = tty
	cmd.Stderr = os.Stderr

	if err = cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run command: %w", err)
	}

	f, err := e.fs.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open temporary file: %w", err)
	}

	return f, nil
}

func (e *cmdEditor) writeInputToFile(msg string) (string, error) {
	f, err := fsutils.CreateTempFile(e.fs, "editor-msg-*.json")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(msg); err != nil {
		return f.Name(), fmt.Errorf("failed to write string: %w", err)
	}

	return f.Name(), nil
}

func (e *cmdEditor) cleanUp(fileName string) {
	if fileName == "" {
		return
	}

	if err := e.fs.Remove(fileName); err != nil {
		// TODO: pass a logger instance instead of calling the global one.
		log.Printf("couldn't remove temporary file: %v", err)
	}
}

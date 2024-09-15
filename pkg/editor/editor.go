package editor

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/afero"
)

// ErrNoCmd is an error returned when no command is provided to run.
var ErrNoCmd = errors.New("no command to run")

// Editor is an interface that allows message editing in an arbitrary text editor.
type Editor interface {
	Run(msg string) (string, error)
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
func (e *cmdEditor) Run(msg string) (string, error) {
	fileName, err := e.writeMsgToFile(msg)
	defer e.cleanUp(fileName)

	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}

	cmdArgs := strings.Split(e.cmd, " ")
	if len(cmdArgs) < 1 {
		return "", ErrNoCmd
	}

	cmdArgs = append(cmdArgs, fileName)

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...) //nolint:gosec // This should be fine if ran on a user machine.
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err = cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run command: %w", err)
	}

	b, err := afero.ReadFile(e.fs, fileName)
	if err != nil {
		return "", fmt.Errorf("failed to read temporary file: %w", err)
	}

	return string(b), nil
}

func (e *cmdEditor) writeMsgToFile(msg string) (string, error) {
	f, err := afero.TempFile(e.fs, afero.GetTempDir(e.fs, ""), "editor-msg-*.txt")
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

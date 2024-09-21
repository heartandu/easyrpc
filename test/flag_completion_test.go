package test

import (
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestProtoFileFlagCompletion(t *testing.T) {
	fs := afero.NewCopyOnWriteFs(afero.NewOsFs(), afero.NewMemMapFs())

	conf, err := createTempFile(fs, "config.yaml", `
        import_paths:
          - `+importPath)
	if err != nil {
		t.Fatalf("failed to create proto files config file: %v", err)
	}

	tests := []struct {
		name          string
		args          []string
		want          []string
		wantDirective string
	}{
		{
			name:          "flag completion",
			args:          []string{"-i", importPath, "-p", ""},
			want:          []string{importPath},
			wantDirective: "ShellCompDirectiveFilterDirs",
		},
		{
			name:          "flag with config completion",
			args:          []string{"--config", conf, "-p", ""},
			want:          []string{importPath},
			wantDirective: "ShellCompDirectiveFilterDirs",
		},
		{
			name:          "no import paths",
			args:          []string{"-p", ""},
			want:          []string{},
			wantDirective: "ShellCompDirectiveDefault",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := runRootCmdAutocomplete(fs, tt.args...)
			if err != nil {
				t.Fatalf("command failed: output = %v, err = %v", string(b), err)
			}

			lines := strings.Split(strings.TrimSpace(string(b)), "\n")
			if len(lines) < 2 {
				t.Fatalf("autocomplete returned unknown response: %v", lines)
			}

			require.Equal(t, tt.want, lines[:len(lines)-2])
			require.Contains(t, lines[len(lines)-1], tt.wantDirective)
		})
	}
}

func TestPackageFlagCompletion(t *testing.T) {
	fs := afero.NewCopyOnWriteFs(afero.NewOsFs(), afero.NewMemMapFs())

	conf, err := createTempFile(fs, "config.yaml", `
        import_paths:
          - `+importPath+`
        proto_files:
          - `+protoFile)
	if err != nil {
		t.Fatalf("failed to create proto files config file: %v", err)
	}

	reflectConf, err := createTempFile(fs, "reflect_conf.yaml", `
        reflection: true
        address: `+address(insecureSocket))
	if err != nil {
		t.Fatalf("failed to create proto files config file: %v", err)
	}

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "empty flag",
			args: []string{"-i", importPath, "-p", protoFile, "--package", ""},
			want: []string{"echo"},
		},
		{
			name: "empty flag with config",
			args: []string{"--config", conf, "--package", ""},
			want: []string{"echo"},
		},
		{
			name: "empty flag reflection",
			args: []string{"-r", "-a", address(insecureSocket), "--package", ""},
			want: []string{"echo", "grpc.reflection.v1", "grpc.reflection.v1alpha"},
		},
		{
			name: "empty flag reflection with config",
			args: []string{"--config", reflectConf, "--package", ""},
			want: []string{"echo", "grpc.reflection.v1", "grpc.reflection.v1alpha"},
		},
		{
			name: "partial complete",
			args: []string{"--config", reflectConf, "--package", "v1"},
			want: []string{"grpc.reflection.v1", "grpc.reflection.v1alpha"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := runRootCmdAutocomplete(fs, tt.args...)
			if err != nil {
				t.Fatalf("command failed: output = %v, err = %v", string(b), err)
			}

			lines := strings.Split(strings.TrimSpace(string(b)), "\n")
			if len(lines) < 2 {
				t.Fatalf("autocomplete returned unknown response: %v", lines)
			}

			require.Equal(t, tt.want, lines[:len(lines)-2])
		})
	}
}

func TestServiceFlagCompletion(t *testing.T) {
	fs := afero.NewCopyOnWriteFs(afero.NewOsFs(), afero.NewMemMapFs())

	conf, err := createTempFile(fs, "config.yaml", `
        import_paths:
          - `+importPath+`
        proto_files:
          - `+protoFile)
	if err != nil {
		t.Fatalf("failed to create proto files config file: %v", err)
	}

	confWithPackage, err := createTempFile(fs, "package_config.yaml", `
        import_paths:
          - `+importPath+`
        proto_files:
          - `+protoFile+`
        package: echo`)
	if err != nil {
		t.Fatalf("failed to create proto files config file: %v", err)
	}

	reflectConf, err := createTempFile(fs, "reflect_conf.yaml", `
        reflection: true
        address: `+address(insecureSocket))
	if err != nil {
		t.Fatalf("failed to create proto files config file: %v", err)
	}

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "empty flag",
			args: []string{"-i", importPath, "-p", protoFile, "--package", "echo", "--service", ""},
			want: []string{"EchoService"},
		},
		{
			name: "empty flag with config",
			args: []string{"--config", conf, "--package", "echo", "--service", ""},
			want: []string{"EchoService"},
		},
		{
			name: "empty flag reflection",
			args: []string{"-r", "-a", address(insecureSocket), "--package", "echo", "--service", ""},
			want: []string{"EchoService"},
		},
		{
			name: "empty flag reflection with config",
			args: []string{"--config", reflectConf, "--package", "echo", "--service", ""},
			want: []string{"EchoService"},
		},
		{
			name: "empty flag without package flag",
			args: []string{"--config", conf, "--service", ""},
			want: []string{"echo.EchoService"},
		},
		{
			name: "empty flag without package flag using reflect",
			args: []string{"--config", reflectConf, "--service", ""},
			want: []string{
				"echo.EchoService",
				"grpc.reflection.v1.ServerReflection",
				"grpc.reflection.v1alpha.ServerReflection",
			},
		},
		{
			name: "config with package",
			args: []string{"--config", confWithPackage, "--service", ""},
			want: []string{"EchoService"},
		},
		{
			name: "partial completion",
			args: []string{"--config", reflectConf, "--service", "Server"},
			want: []string{
				"grpc.reflection.v1.ServerReflection",
				"grpc.reflection.v1alpha.ServerReflection",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := runRootCmdAutocomplete(fs, tt.args...)
			if err != nil {
				t.Fatalf("command failed: output = %v, err = %v", string(b), err)
			}

			lines := strings.Split(strings.TrimSpace(string(b)), "\n")
			if len(lines) < 2 {
				t.Fatalf("autocomplete returned unknown response: %v", lines)
			}

			require.Equal(t, tt.want, lines[:len(lines)-2])
		})
	}
}

func runRootCmdAutocomplete(fs afero.Fs, args ...string) ([]byte, error) {
	return run(fs, nil, append([]string{"__complete"}, args...)...)
}

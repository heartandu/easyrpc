package config_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/heartandu/easyrpc/internal/config"
)

const editorVar = "EDITOR"

// setupTestCommand creates a cobra.Command with all config flags registered.
// This is called once and reused across tests since flags are not modified.
func setupTestCommand() *cobra.Command {
	cmd := &cobra.Command{}
	flags := cmd.PersistentFlags()

	// Proto flags
	flags.StringSlice("import-path", nil, "")
	flags.Bool("import-all", false, "")
	flags.StringSlice("proto-file", nil, "")

	// Server flags
	flags.String("address", "", "")
	flags.Bool("reflection", false, "")
	flags.Bool("web", false, "")

	// TLS flags
	flags.Bool("tls", false, "")
	flags.String("cacert", "", "")
	flags.String("cert", "", "")
	flags.String("key", "", "")

	// Request flags
	flags.StringSlice("metadata", nil, "")
	flags.String("package", "", "")
	flags.String("service", "", "")

	return cmd
}

//nolint:tparallel // test sets up environment variables
func TestDecoder_Decode(t *testing.T) {
	t.Setenv(editorVar, "")

	tests := []struct {
		name      string
		files     []string
		setupFs   func(afero.Fs)
		flagSetup func(*cobra.Command)
		want      config.Config
		wantErr   bool
	}{
		{
			name:  "all configuration values from file",
			files: []string{"config.yaml"},
			setupFs: func(fs afero.Fs) {
				content := `
import_paths:
  - /path/to/proto
  - /another/path
import_all: true
proto_files:
  - service.proto
  - another.proto
address: localhost:8080
reflection: true
web: true
tls: true
cacert: /path/to/ca.crt
cert: /path/to/client.crt
key: /path/to/client.key
metadata:
  Authorization: Bearer token
  X-Custom-Header: custom-value
package: test.v1
service: TestService
`
				require.NoError(t, afero.WriteFile(fs, "config.yaml", []byte(content), 0o644))
			},
			flagSetup: nil,
			want: config.Config{
				Proto: config.Proto{
					ImportPaths: []string{"/path/to/proto", "/another/path"},
					ImportAll:   true,
					ProtoFiles:  []string{"service.proto", "another.proto"},
				},
				Server: config.Server{
					Address:    "localhost:8080",
					Reflection: true,
					Web:        true,
				},
				TLS: config.TLS{
					Enabled: true,
					CACert:  "/path/to/ca.crt",
					Cert:    "/path/to/client.crt",
					Key:     "/path/to/client.key",
				},
				Request: config.Request{
					Metadata: map[string]string{
						"Authorization":   "Bearer token",
						"X-Custom-Header": "custom-value",
					},
					Package: "test.v1",
					Service: "TestService",
				},
			},
			wantErr: false,
		},
		{
			name:  "all configuration values from flags",
			files: []string{},
			setupFs: func(fs afero.Fs) {
				// No files
			},
			flagSetup: func(cmd *cobra.Command) {
				require.NoError(t, cmd.PersistentFlags().Set("import-path", "/path/to/proto"))
				require.NoError(t, cmd.PersistentFlags().Set("import-path", "/another/path"))
				require.NoError(t, cmd.PersistentFlags().Set("import-all", "true"))
				require.NoError(t, cmd.PersistentFlags().Set("proto-file", "service.proto"))
				require.NoError(t, cmd.PersistentFlags().Set("proto-file", "another.proto"))
				require.NoError(t, cmd.PersistentFlags().Set("address", "localhost:8080"))
				require.NoError(t, cmd.PersistentFlags().Set("reflection", "true"))
				require.NoError(t, cmd.PersistentFlags().Set("web", "true"))
				require.NoError(t, cmd.PersistentFlags().Set("tls", "true"))
				require.NoError(t, cmd.PersistentFlags().Set("cacert", "/path/to/ca.crt"))
				require.NoError(t, cmd.PersistentFlags().Set("cert", "/path/to/client.crt"))
				require.NoError(t, cmd.PersistentFlags().Set("key", "/path/to/client.key"))
				require.NoError(t, cmd.PersistentFlags().Set("metadata", "Authorization: Bearer token"))
				require.NoError(t, cmd.PersistentFlags().Set("metadata", "X-Custom-Header: custom-value"))
				require.NoError(t, cmd.PersistentFlags().Set("package", "test.v1"))
				require.NoError(t, cmd.PersistentFlags().Set("service", "TestService"))
			},
			want: config.Config{
				Proto: config.Proto{
					ImportPaths: []string{"/path/to/proto", "/another/path"},
					ImportAll:   true,
					ProtoFiles:  []string{"service.proto", "another.proto"},
				},
				Server: config.Server{
					Address:    "localhost:8080",
					Reflection: true,
					Web:        true,
				},
				TLS: config.TLS{
					Enabled: true,
					CACert:  "/path/to/ca.crt",
					Cert:    "/path/to/client.crt",
					Key:     "/path/to/client.key",
				},
				Request: config.Request{
					Metadata: map[string]string{
						"Authorization":   "Bearer token",
						"X-Custom-Header": "custom-value",
					},
					Package: "test.v1",
					Service: "TestService",
				},
			},
			wantErr: false,
		},
		{
			name:      "no config files",
			files:     []string{},
			setupFs:   nil,
			flagSetup: nil,
			want:      config.Config{},
			wantErr:   false,
		},
		{
			name:      "config file not found is skipped",
			files:     []string{"nonexistent.yaml"},
			setupFs:   nil,
			flagSetup: nil,
			want:      config.Config{},
			wantErr:   false,
		},
		{
			name:  "invalid yaml syntax",
			files: []string{"invalid.yaml"},
			setupFs: func(fs afero.Fs) {
				content := `
import_paths: [
  - /path/to/proto
invalid yaml: [:
`
				require.NoError(t, afero.WriteFile(fs, "invalid.yaml", []byte(content), 0o644))
			},
			flagSetup: nil,
			want:      config.Config{},
			wantErr:   true,
		},
		{
			name:  "multiple config files merged",
			files: []string{"config1.yaml", "config2.yaml"},
			setupFs: func(fs afero.Fs) {
				content1 := `
import_paths:
  - /path/to/proto
address: localhost:8080
reflection: true
`
				content2 := `
address: localhost:9090
web: true
metadata:
  key1: value1
  key2: value2
`
				require.NoError(t, afero.WriteFile(fs, "config1.yaml", []byte(content1), 0o644))
				require.NoError(t, afero.WriteFile(fs, "config2.yaml", []byte(content2), 0o644))
			},
			flagSetup: nil,
			want: config.Config{
				Proto: config.Proto{
					ImportPaths: []string{"/path/to/proto"},
				},
				Server: config.Server{
					Address:    "localhost:9090", // Overridden by config2
					Reflection: true,
					Web:        true,
				},
				Request: config.Request{
					Metadata: map[string]string{"key1": "value1", "key2": "value2"},
				},
			},
			wantErr: false,
		},
		{
			name:  "flags override config file values",
			files: []string{"config.yaml"},
			setupFs: func(fs afero.Fs) {
				content := `
address: localhost:8080
reflection: true
tls: false
`
				require.NoError(t, afero.WriteFile(fs, "config.yaml", []byte(content), 0o644))
			},
			flagSetup: func(cmd *cobra.Command) {
				require.NoError(t, cmd.PersistentFlags().Set("address", "localhost:9090"))
				require.NoError(t, cmd.PersistentFlags().Set("tls", "true"))
			},
			want: config.Config{
				Server: config.Server{
					Address:    "localhost:9090", // From flag
					Reflection: true,             // From file
				},
				TLS: config.TLS{
					Enabled: true, // From flag, overrides file
				},
			},
			wantErr: false,
		},
		{
			name:  "nested map merge",
			files: []string{"config1.yaml", "config2.yaml"},
			setupFs: func(fs afero.Fs) {
				content1 := `
metadata:
  key1: value1
  key2: value2
`
				content2 := `
metadata:
  key2: overwritten
  key3: value3
`
				require.NoError(t, afero.WriteFile(fs, "config1.yaml", []byte(content1), 0o644))
				require.NoError(t, afero.WriteFile(fs, "config2.yaml", []byte(content2), 0o644))
			},
			flagSetup: func(cmd *cobra.Command) {
				require.NoError(t, cmd.PersistentFlags().Set("metadata", "key3: overwritten"))
			},
			want: config.Config{
				Request: config.Request{
					Metadata: map[string]string{
						"key1": "value1",
						"key2": "overwritten",
						"key3": "overwritten",
					},
				},
			},
			wantErr: false,
		},
		{
			name:  "empty yaml file returns error",
			files: []string{"empty.yaml"},
			setupFs: func(fs afero.Fs) {
				require.NoError(t, afero.WriteFile(fs, "empty.yaml", []byte(""), 0o644))
			},
			flagSetup: nil,
			want:      config.Config{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a fresh MemMapFs for each test
			fs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}

			// Create a fresh command for each test to avoid flag state pollution
			cmd := setupTestCommand()
			if tt.flagSetup != nil {
				tt.flagSetup(cmd)
			}

			decoder := config.NewDecoder(fs, cmd, tt.files)
			got, err := decoder.Decode()

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestDecoder_Decode_EditorEnvVar(t *testing.T) {
	tests := []struct {
		name   string
		envVar string
		setEnv bool
		want   config.Config
	}{
		{
			name:   "EDITOR env var is set",
			envVar: "vim",
			setEnv: true,
			want: config.Config{
				Editor: config.Editor{
					Cmd: "vim",
				},
			},
		},
		{
			name:   "EDITOR env var is not set",
			envVar: "",
			setEnv: false,
			want: config.Config{
				Editor: config.Editor{
					Cmd: "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()

			envVar := ""
			if tt.setEnv {
				envVar = tt.envVar
			}

			t.Setenv(editorVar, envVar)

			decoder := config.NewDecoder(fs, nil, nil)
			got, err := decoder.Decode()

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

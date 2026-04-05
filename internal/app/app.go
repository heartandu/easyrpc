package app

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/heartandu/easyrpc/internal/autocomplete"
	"github.com/heartandu/easyrpc/internal/config"
)

const (
	defaultConfigName = ".easyrpc.yaml"

	flagConfig     = "config"
	flagAddress    = "address"
	flagImportPath = "import-path"
	flagImportAll  = "import-all"
	flagProtoFile  = "proto-file"
	flagReflection = "reflection"
	flagWeb        = "web"
	flagTLS        = "tls"
	flagCACert     = "cacert"
	flagCert       = "cert"
	flagKey        = "key"
	flagPackage    = "package"
	flagService    = "service"
	flagMetadata   = "metadata"
)

// App is a container of all application initialization and logic.
type App struct {
	version string

	cfgFile string
	cfg     config.Config

	fs     afero.Fs
	cmd    *cobra.Command
	pflags *pflag.FlagSet
}

// NewApp returns a new instance of App.
func NewApp(version string) *App {
	cmd := &cobra.Command{
		Use:   "easyrpc",
		Short: "An easy gRPC client",
		Long: `easyrpc is a CLI and REPL uitility to make gRPC or gRPC-Web calls.
The main purpose of this utility is for manual API testing.`,
		SilenceUsage: true,
	}

	return &App{
		version: version,
		fs:      afero.NewOsFs(),
		cmd:     cmd,
		pflags:  cmd.PersistentFlags(),
	}
}

// SetOutput sets output writer for all commands.
func (a *App) SetOutput(w io.Writer) {
	a.cmd.SetOut(w)
	a.cmd.SetErr(w)
}

// SetInput sets input reader for all commands.
func (a *App) SetInput(r io.Reader) {
	a.cmd.SetIn(r)
}

// SetFs sets a filesystem wrapper.
func (a *App) SetFs(fs afero.Fs) {
	a.fs = fs
}

// Run sets up an application and executes the command.
func (a *App) Run() error {
	a.bindPFlags()
	a.registerCommands()

	cobra.OnInitialize(a.onInit)

	return a.cmd.Execute() //nolint:wrapcheck // It's not informative to wrap the error here.
}

// bindPFlags sets up application global flags.
func (a *App) bindPFlags() {
	protoCompletion := autocomplete.NewProtoComp(a.fs, a.readConfig)
	protoFileCompletion := autocomplete.NewProtoFileFlag(a.readConfig)

	a.pflags.StringVar(&a.cfgFile, flagConfig, "", "config file (default is $HOME/.easyrpc.yaml or ./.easyrpc.yaml)")
	a.pflags.StringP(flagAddress, "a", "", `remote host address in format "host:port" or "host:port/prefix"`)
	a.pflags.StringSliceP(
		flagImportPath,
		"i",
		nil,
		"proto import path, can provide multiple paths by repeating the flag",
	)
	a.pflags.Bool(flagImportAll, false, "import all proto files from import path")
	a.pflags.StringSliceP(
		flagProtoFile,
		"p",
		nil,
		"proto files to use, can provide multiple files by repeating the flag",
	)
	a.cmd.RegisterFlagCompletionFunc(flagProtoFile, protoFileCompletion.Complete)
	a.pflags.BoolP(flagReflection, "r", false, "use server reflection to make requests")
	a.pflags.BoolP(flagWeb, "w", false, "use gRPC-Web client to make requests")
	a.pflags.Bool(flagTLS, false, "use a secure TLS connection")
	a.pflags.String(flagCACert, "", "CA certificate file for verifying the server")
	a.pflags.String(flagCert, "", "certificate file for mutual TLS auth. It must be provided along with --key")
	a.pflags.String(flagKey, "", "private key for mutual TLS auth. It must be provided along with --cert")
	a.pflags.String(flagPackage, "", "the package name to use as default")
	a.cmd.RegisterFlagCompletionFunc(flagPackage, protoCompletion.CompletePackage)
	a.pflags.String(flagService, "", "the service name to use as default")
	a.cmd.RegisterFlagCompletionFunc(flagService, protoCompletion.CompleteService)
	a.pflags.StringSliceP(
		flagMetadata,
		"H",
		nil,
		`metadata (headers) that are attached to every request in format "key: value"`,
	)
}

// registerCommands adds all application commands to the root one.
func (a *App) registerCommands() {
	a.registerVersionCmd()
	a.registerCallCmd()
	a.registerRequestCmd()
	a.registerConfigCmd()
}

func (a *App) onInit() {
	var err error

	a.cfg, err = a.readConfig()
	cobra.CheckErr(err)
}

// readConfig reads in config file and ENV variables if set.
func (a *App) readConfig() (config.Config, error) {
	// Find home directory.
	home, err := os.UserHomeDir()
	if err != nil {
		return config.Config{}, fmt.Errorf("failedt to read user home directory: %w", err)
	}

	files := []string{
		path.Join(home, defaultConfigName),
		path.Join(".", defaultConfigName),
	}

	// Use config file from the flag.
	if a.cfgFile != "" {
		files = append(files, a.cfgFile)
	}

	cfg, err := config.NewDecoder(a.fs, a.cmd, files).Decode()
	if err != nil {
		return config.Config{}, fmt.Errorf("failed to decode config: %w", err)
	}

	return cfg, nil
}

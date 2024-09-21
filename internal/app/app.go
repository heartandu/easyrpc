package app

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/heartandu/easyrpc/internal/autocomplete"
	"github.com/heartandu/easyrpc/internal/config"
)

const (
	defaultConfigName = ".easyrpc.yaml"

	flagConfig     = "config"
	flagAddress    = "address"
	flagImportPath = "import-path"
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
	cfgFile string
	cfg     config.Config

	fs     afero.Fs
	cmd    *cobra.Command
	viper  *viper.Viper
	pflags *pflag.FlagSet
}

// NewApp returns a new instance of App.
func NewApp() *App {
	cmd := &cobra.Command{
		Use:   "easyrpc",
		Short: "An easy gRPC client",
		Long: `easyrpc is a CLI and REPL uitility to make gRPC or gRPC-Web calls.
The main purpose of this utility is for manual API testing.`,
		SilenceUsage: true,
	}

	return &App{
		fs:     afero.NewOsFs(),
		cmd:    cmd,
		viper:  viper.New(),
		pflags: cmd.PersistentFlags(),
	}
}

// SetOutput sets output writer for all commands.
func (a *App) SetOutput(w io.Writer) {
	a.cmd.SetOut(w)
	a.cmd.SetOutput(w)
}

// SetInput sets input reader for all commands.
func (a *App) SetInput(r io.Reader) {
	a.cmd.SetIn(r)
}

// SetFs sets a filesystem wrapper.
func (a *App) SetFs(fs afero.Fs) { //nolint:gocritic,revive // The scope is small enough to afford such shadowing.
	a.fs = fs
	a.viper.SetFs(fs)
}

// Run sets up an application and executes the command.
func (a *App) Run() error {
	a.bindPFlags()
	a.bindPFlagsToConfig()
	a.bindEnv()
	a.registerCommands()

	cobra.OnInitialize(a.readConfig)

	return a.cmd.Execute() //nolint:wrapcheck // It's not informative to wrap the error here.
}

// bindPFlags sets up application global flags.
func (a *App) bindPFlags() {
	a.pflags.StringVar(&a.cfgFile, flagConfig, "", "config file (default is $HOME/.easyrpc.yaml or ./.easyrpc.yaml)")
	a.pflags.StringP(flagAddress, "a", "", `remote host address in format "host:port" or "host:port/prefix"`)
	a.pflags.StringSliceP(
		flagImportPath,
		"i",
		nil,
		"proto import path, can provide multiple paths by repeating the flag",
	)
	a.pflags.StringSliceP(
		flagProtoFile,
		"p",
		nil,
		"proto files to use, can provide multiple files by repeating the flag",
	)
	a.cmd.RegisterFlagCompletionFunc(
		flagProtoFile,
		autocomplete.NewProtoFileFlag(a.viper).Complete,
	)
	a.pflags.BoolP(flagReflection, "r", false, "use server reflection to make requests")
	a.pflags.BoolP(flagWeb, "w", false, "use gRPC-Web client to make requests")
	a.pflags.Bool(flagTLS, false, "use a secure TLS connection")
	a.pflags.String(flagCACert, "", "CA certificate file for verifying the server")
	a.pflags.String(flagCert, "", "certificate file for mutual TLS auth. It must be provided along with --key")
	a.pflags.String(flagKey, "", "private key for mutual TLS auth. It must be provided along with --cert")
	a.pflags.String(flagPackage, "", "the package name to use as default")
	a.cmd.RegisterFlagCompletionFunc(
		flagPackage,
		autocomplete.NewProtoComp(a.fs, a.viper).CompletePackage,
	)
	a.pflags.String(flagService, "", "the service name to use as default")
	a.cmd.RegisterFlagCompletionFunc(
		flagService,
		autocomplete.NewProtoComp(a.fs, a.viper).CompleteService,
	)
	a.pflags.StringToStringP(flagMetadata, "H", nil, "default headers that are attached to every request")
}

// bindPFlagsToConfig binds application global flags to configuration structure.
func (a *App) bindPFlagsToConfig() {
	a.viper.BindPFlag("cacert", a.pflags.Lookup(flagCACert))
	a.viper.BindPFlag("cert", a.pflags.Lookup(flagCert))
	a.viper.BindPFlag("key", a.pflags.Lookup(flagKey))
	a.viper.BindPFlag("address", a.pflags.Lookup(flagAddress))
	a.viper.BindPFlag("reflection", a.pflags.Lookup(flagReflection))
	a.viper.BindPFlag("web", a.pflags.Lookup(flagWeb))
	a.viper.BindPFlag("tls", a.pflags.Lookup(flagTLS))
	a.viper.BindPFlag("import_paths", a.pflags.Lookup(flagImportPath))
	a.viper.BindPFlag("proto_files", a.pflags.Lookup(flagProtoFile))
	a.viper.BindPFlag("package", a.pflags.Lookup(flagPackage))
	a.viper.BindPFlag("service", a.pflags.Lookup(flagService))
	a.viper.BindPFlag("metadata", a.pflags.Lookup(flagMetadata))
}

func (a *App) bindEnv() {
	a.viper.BindEnv("editor")
}

// registerCommands adds all application commands to the root one.
func (a *App) registerCommands() {
	a.registerVersionCmd()
	a.registerCallCmd()
	a.registerRequestCmd()
	a.registerConfigCmd()
}

// readConfig reads in config file and ENV variables if set.
func (a *App) readConfig() {
	// Find home directory.
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	a.viper.SetEnvPrefix("easyrpc")
	a.viper.AutomaticEnv() // read in environment variables that match

	files := []string{
		path.Join(home, defaultConfigName),
		path.Join(".", defaultConfigName),
	}

	// Use config file from the flag.
	if a.cfgFile != "" {
		files = append(files, a.cfgFile)
	}

	var (
		notFoundErr viper.ConfigFileNotFoundError
		fsErr       *fs.PathError
	)

	for _, file := range files {
		a.viper.SetConfigFile(file)

		if err := a.viper.MergeInConfig(); err != nil && !errors.As(err, &notFoundErr) && !errors.As(err, &fsErr) {
			cobra.CheckErr(fmt.Errorf("failed to read config: %w", err))
		}
	}

	if err := a.viper.Unmarshal(&a.cfg); err != nil {
		cobra.CheckErr(fmt.Errorf("failed to unmarshal config: %w", err))
	}
}

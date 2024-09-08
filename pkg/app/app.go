package app

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/heartandu/easyrpc/pkg/config"
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

func (a *App) SetFs(fs afero.Fs) {
	a.fs = fs
	a.viper.SetFs(fs)
}

// Run sets up an application and executes the command.
func (a *App) Run() error {
	a.bindPFlags()
	a.bindPFlagsToConfig()
	a.registerCommands()

	cobra.OnInitialize(a.readConfig)

	return a.cmd.Execute() //nolint:wrapcheck // It's not informative to wrap the error here.
}

// bindPFlags sets up application global flags.
func (a *App) bindPFlags() {
	a.pflags.StringVar(&a.cfgFile, "config", "", "config file (default is $HOME/.easyrpc.yaml or ./.easyrpc.yaml)")
	a.pflags.StringP("address", "a", "", `remote host address in format "host:port" or "host:port/prefix"`)
	a.pflags.StringSliceP(
		"import-path",
		"i",
		nil,
		"proto import path, can provide multiple paths by repeating the flag",
	)
	a.pflags.StringSliceP(
		"proto-file",
		"p",
		nil,
		"proto files to use, can provide multiple files by repeating the flag",
	)
	a.pflags.BoolP("reflection", "r", false, "use server reflection to make requests")
	a.pflags.Bool("tls", false, "use a secure TLS connection")
	a.pflags.String("cacert", "", "CA certificate file for verifying the server")
	a.pflags.String("cert", "", "the certificate file for mutual TLS auth. It must be provided with --cert-key")
	a.pflags.String("cert-key", "", "the private key for mutual TLS auth. It must be provided with --cert")
	a.pflags.String("package", "", "the package name to use as default")
	a.pflags.String("service", "", "the service name to use as default")
	a.pflags.StringToStringP("metadata", "H", nil, "default headers that are attached to every request")
}

// bindPFlagsToConfig binds application global flags to configuration structure.
func (a *App) bindPFlagsToConfig() {
	a.viper.BindPFlag("cacert", a.pflags.Lookup("cacert"))            //nolint:errcheck // viper flag bind
	a.viper.BindPFlag("cert", a.pflags.Lookup("cert"))                //nolint:errcheck // viper flag bind
	a.viper.BindPFlag("cert_key", a.pflags.Lookup("cert-key"))        //nolint:errcheck // viper flag bind
	a.viper.BindPFlag("address", a.pflags.Lookup("address"))          //nolint:errcheck // viper flag bind
	a.viper.BindPFlag("reflection", a.pflags.Lookup("reflection"))    //nolint:errcheck // viper flag bind
	a.viper.BindPFlag("tls", a.pflags.Lookup("tls"))                  //nolint:errcheck // viper flag bind
	a.viper.BindPFlag("import_paths", a.pflags.Lookup("import-path")) //nolint:errcheck // viper flag bind
	a.viper.BindPFlag("proto_files", a.pflags.Lookup("proto-file"))   //nolint:errcheck // viper flag bind
	a.viper.BindPFlag("package", a.pflags.Lookup("package"))          //nolint:errcheck // viper flag bind
	a.viper.BindPFlag("service", a.pflags.Lookup("service"))          //nolint:errcheck // viper flag bind
	a.viper.BindPFlag("metadata", a.pflags.Lookup("metadata"))        //nolint:errcheck // viper flag bind
}

// registerCommands adds all application commands to the root one.
func (a *App) registerCommands() {
	a.registerVersionCmd()
	a.registerCallCmd()
}

// readConfig reads in config file and ENV variables if set.
func (a *App) readConfig() {
	if a.cfgFile != "" {
		// Use config file from the flag.
		a.viper.SetConfigFile(a.cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home or current working directory with name ".easyrpc" (without extension).
		a.viper.AddConfigPath(".")
		a.viper.AddConfigPath(home)
		a.viper.SetConfigType("yaml")
		a.viper.SetConfigName(".easyrpc")
	}

	a.viper.AutomaticEnv() // read in environment variables that match

	var notFoundErr viper.ConfigFileNotFoundError
	if err := a.viper.ReadInConfig(); err != nil && !errors.As(err, &notFoundErr) {
		cobra.CheckErr(fmt.Errorf("failed to read config: %w", err))
	}

	if err := a.viper.Unmarshal(&a.cfg); err != nil {
		cobra.CheckErr(fmt.Errorf("failed to unmarshal config: %w", err))
	}
}

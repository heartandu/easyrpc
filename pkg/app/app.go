package app

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/heartandu/easyrpc/pkg/config"
)

// App is a container of all application initialization and logic.
type App struct {
	cfgFile string
	cfg     config.Config

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
	}

	return &App{
		cmd:    cmd,
		viper:  viper.New(),
		pflags: cmd.PersistentFlags(),
	}
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
}

// bindPFlagsToConfig binds application global flags to configuration structure.
func (a *App) bindPFlagsToConfig() {
	a.viper.BindPFlag("server.address", a.pflags.Lookup("address")) //nolint:errcheck // viper flag bind
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
		cobra.CheckErr(err)
	}

	cobra.CheckErr(a.viper.Unmarshal(&a.cfg))
}

package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func (a *App) registerConfigCmd() {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration files manipulation",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "init [PATH]",
			Short: "Populate config file with current or default values",
			Long:  "The command populates currently loaded configuration to the PATH or to the current directory",
			RunE: func(_ *cobra.Command, args []string) error {
				cfgPath := defaultConfigName

				if len(args) > 0 {
					cfgPath = args[0]
				}

				if err := a.viper.WriteConfigAs(cfgPath); err != nil {
					return fmt.Errorf("failed to write config: %w", err)
				}

				return nil
			},
		},
		&cobra.Command{
			Use:   "dump",
			Short: "Dump current configuration to stdout",
			RunE: func(cmd *cobra.Command, _ []string) error {
				e := yaml.NewEncoder(cmd.OutOrStdout())
				if err := e.Encode(a.viper.AllSettings()); err != nil {
					return fmt.Errorf("failed to marshal settings: %w", err)
				}

				return nil
			},
		},
	)

	a.cmd.AddCommand(cmd)
}

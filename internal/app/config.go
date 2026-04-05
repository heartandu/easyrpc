package app

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var ErrConfigAlreadyExists = errors.New("config file already exists")

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

				file, err := a.fs.OpenFile(cfgPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o666)
				if err != nil {
					if errors.Is(err, os.ErrExist) {
						return fmt.Errorf("%w: %s", ErrConfigAlreadyExists, cfgPath)
					}

					return fmt.Errorf("failed to create config file: %w", err)
				}
				defer file.Close()

				enc := yaml.NewEncoder(file)
				if err := enc.Encode(a.cfg); err != nil {
					return fmt.Errorf("failed to encode config: %w", err)
				}

				return nil
			},
		},
		&cobra.Command{
			Use:   "dump",
			Short: "Dump current configuration to stdout",
			RunE: func(cmd *cobra.Command, _ []string) error {
				e := yaml.NewEncoder(cmd.OutOrStdout())
				if err := e.Encode(a.cfg); err != nil {
					return fmt.Errorf("failed to marshal settings: %w", err)
				}

				return nil
			},
		},
	)

	a.cmd.AddCommand(cmd)
}

package config

import (
	"fmt"
	"time"

	"github.com/fatih/color"

	"github.com/AlecAivazis/survey/v2"
	"github.com/candrewlee14/webman/cmd/config/add"
	"github.com/candrewlee14/webman/cmd/config/remove"
	"github.com/candrewlee14/webman/config"
	"github.com/spf13/cobra"
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "subcommands for webman config",
	Long: `

The "config" subcommand allows you to change your base webman config.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		q := &survey.Input{
			Message: "Refresh interval",
			Help:    "How long before webman should refresh a package repository",
			Default: cfg.RefreshInterval.String(),
		}

		if err := survey.AskOne(q, &cfg.RefreshInterval, survey.WithValidator(func(ans interface{}) error {
			_, err := time.ParseDuration(fmt.Sprint(ans))
			return err
		})); err != nil {
			return err
		}

		if err := cfg.Save(); err != nil {
			return err
		}

		color.HiGreen("Config saved")
		return nil
	},
}

func init() {
	ConfigCmd.AddCommand(add.AddCmd)
	ConfigCmd.AddCommand(remove.RemoveCmd)
}

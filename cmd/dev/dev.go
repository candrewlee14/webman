package dev

import (
	"github.com/candrewlee14/webman/cmd/dev/bintest"
	"github.com/candrewlee14/webman/cmd/dev/check"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var DevCmd = &cobra.Command{
	Use:   "dev",
	Short: "subcommands for package maintainers",
	Long: `

The "dev" subcommand helps package maintainers check their changes.
`,
}

func init() {
	DevCmd.AddCommand(check.CheckCmd)
	DevCmd.AddCommand(bintest.BintestCmd)
}

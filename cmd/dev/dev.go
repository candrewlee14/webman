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
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },

}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	DevCmd.AddCommand(check.CheckCmd)
	DevCmd.AddCommand(bintest.BintestCmd)
}

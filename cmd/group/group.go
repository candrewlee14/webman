package group

import (
	groupadd "github.com/candrewlee14/webman/cmd/group/add"
	groupremove "github.com/candrewlee14/webman/cmd/group/remove"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var GroupCmd = &cobra.Command{
	Use:   "group",
	Short: "manage groups of packages",
	Long: `

The "group" subcommand manages a group of packages.
`,
	Example: `
webman group add
webman group remove
`, // Uncomment the following line if your bare application
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

	GroupCmd.AddCommand(groupadd.AddCmd)
	GroupCmd.AddCommand(groupremove.RemoveCmd)
}

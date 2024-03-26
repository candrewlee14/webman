package group

import (
	groupadd "github.com/candrewlee14/webman/cmd/group/add"
	groupremove "github.com/candrewlee14/webman/cmd/group/remove"
	groupsearch "github.com/candrewlee14/webman/cmd/group/search"

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
`,
}

func init() {
	GroupCmd.AddCommand(groupadd.AddCmd)
	GroupCmd.AddCommand(groupremove.RemoveCmd)
	GroupCmd.AddCommand(groupsearch.SearchCmd)
}

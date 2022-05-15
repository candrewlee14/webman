package version

import (
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
	BuiltBy = "unknown"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "display the webman version",
	Long: `
The "version" subcommand displays the latest webman version.`,
	Example: `webman version`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 0 {
			cmd.Help()
			os.Exit(0)
		}
		color.Cyan("webman (v%s)", Version)
		color.Yellow("Commit %s", Commit[:8])
		color.Magenta("Built on %s by %s", Date[:10], BuiltBy)
		color.HiBlack("Created by candrewlee14")
	},
}

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
		color.Cyan("webman (%s)", Version)
		verLen := 8
		if len(Commit) < 8 {
			verLen = len(Commit)
		}
		color.Yellow("Commit %s", Commit[:verLen])
		dateLen := 10
		if len(Date) < 10 {
			dateLen = len(Date)
		}
		color.Magenta("Built on %s by %s", Date[:dateLen], BuiltBy)
		color.HiBlack("Created by candrewlee14")
	},
}

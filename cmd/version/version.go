package version

import (
	"runtime/debug"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
	dirty   bool
	Date    = "unknown"
	BuiltBy = "unknown"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "display the webman version",
	Long: `
The "version" subcommand displays the latest webman version.`,
	Example: `webman version`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return cmd.Help()
		}
		color.Cyan("webman (%s)", Version)
		verLen := 8
		if len(Commit) < 8 {
			verLen = len(Commit)
		}
		Commit = Commit[:verLen]
		if dirty {
			Commit += " (with changes)"
		}
		color.Yellow("Commit %s", Commit)
		dateLen := 10
		if len(Date) < 10 {
			dateLen = len(Date)
		}
		color.Magenta("Built on %s by %s", Date[:dateLen], BuiltBy)
		color.HiBlack("Created by candrewlee14")
		return nil
	},
}

func init() {
	if Version != "dev" {
		return
	}
	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				Commit = setting.Value
			case "vcs.time":
				Date = setting.Value
			case "vcs.modified":
				dirty = setting.Value == "true"
			}
		}
	}
}

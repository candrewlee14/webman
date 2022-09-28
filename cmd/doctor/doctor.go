package doctor

import (
	"github.com/candrewlee14/webman/cmd/doctor/check"
	"github.com/candrewlee14/webman/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	fix    bool
	checks = []check.Check{
		check.NestedRecipe,
		check.WindowsSymlink,
	}
)

var DoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "subcommands for webman doctor",
	Long: `

The "doctor" subcommand checks for potential issues. webman can attempt to automatically fix issues using --fix
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		for _, check := range checks {
			color.HiBlue("== %s", check.Name)
			if err := check.Func(cfg, fix); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	DoctorCmd.Flags().BoolVar(&fix, "fix", false, "attempt to fix issues rather than just reporting them")
}

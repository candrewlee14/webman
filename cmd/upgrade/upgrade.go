package upgrade

import (
	"os"

	"github.com/candrewlee14/webman/cmd/add"
	"github.com/candrewlee14/webman/config"
	"github.com/candrewlee14/webman/utils"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	doRefresh bool
)

// upgradeCmd represents the upgrade command
var UpgradeCmd = &cobra.Command{
	Use:   "upgrade [pkgs...]",
	Short: "upgrade packages",
	Long: `
The "upgrade" subcommand adds the latest version of packages, switches to use that version, and removes the old.`,
	Example: `webman upgrade go
webman upgrade go@18.0.0
webman upgrade go zig rg
webman upgrade go@18.0.0 zig@9.1.0 rg@13.0.0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		defer os.RemoveAll(utils.WebmanTmpDir)
		// if local recipe flag is not set
		if utils.RecipeDirFlag == "" {
			// only refresh if not using local
			for _, pkgRepo := range cfg.PkgRepos {
				shouldRefresh, err := pkgRepo.ShouldRefreshRecipes(cfg.RefreshInterval)
				if err != nil {
					return err
				}
				if shouldRefresh || doRefresh {
					color.HiBlue("Refreshing package recipes for %q...", pkgRepo.Name)
					if err = pkgRepo.RefreshRecipes(); err != nil {
						color.Red("%v", err)
					}
				}
			}
		}
		pkgs := add.InstallAllPkgs(cfg.PkgRepos, args, true, true)
		if len(args) != len(pkgs) {
			color.Red("Not all packages installed successfully")
		}
		if len(args) == len(pkgs) {
			color.Green("All %d packages are installed!", len(args))
		}

		return nil
	},
}

func init() {
	UpgradeCmd.Flags().BoolVar(&doRefresh, "refresh", false, "force refresh of package recipes")
}

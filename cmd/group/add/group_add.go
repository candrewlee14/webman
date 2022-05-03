package add

import (
	"fmt"
	"os"
	"webman/cmd/add"
	"webman/multiline"
	"webman/pkgparse"
	"webman/utils"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var doRefresh bool

var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "install a group of packages",
	Long: `

The "group add" subcommand installs a group of packages.
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		utils.Init()
		if len(args) != 1 {
			color.Red("Expected a single package group name")
			cmd.Help()
			os.Exit(1)
		}
		if utils.RecipeDirFlag == "" {
			// only refresh if not using local
			shouldRefresh, err := pkgparse.ShouldRefreshRecipes()
			if err != nil {
				panic(err)
			}
			if shouldRefresh || doRefresh {
				color.HiBlue("Refreshing package recipes...")
				if err = pkgparse.RefreshRecipes(); err != nil {
					fmt.Println(err)
				} else {
					color.HiBlue("%s%sRefreshed package recipes!",
						multiline.MoveUp, multiline.ClearLine)
				}
			}
		}
		group := args[0]
		groupConf := pkgparse.ParseGroupConfig(group)

		if !add.InstallAllPkgs(groupConf.Packages) {
			color.Magenta("Not all packages installed successfully")
			os.Exit(1)
		}
		color.Green("All packages installed successfully from group %s", color.YellowString(group))
	},
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.webman.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	AddCmd.Flags().BoolVar(&doRefresh, "refresh", false, "force refresh of package recipes")

}

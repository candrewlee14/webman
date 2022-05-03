package remove

import (
	"fmt"
	"os"
	"webman/cmd/remove"
	"webman/pkgparse"
	"webman/utils"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var allFlag bool

var RemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove a group of packages",
	Long: `

The "group remove" subcommand removes a group of packages.
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		utils.Init()
		if len(args) != 1 {
			cmd.Help()
			os.Exit(1)
		}
		group := args[0]
		if !allFlag {
			color.Red("Please pass --all flag if you want to remove all versions of all packages in the group " + color.YellowString(group))
			os.Exit(1)
		}
		groupConf := pkgparse.ParseGroupConfig(group)
		for _, pkg := range groupConf.Packages {
			pkgConf, err := pkgparse.ParsePkgConfigLocal(pkg, false)
			if err != nil {
				color.Red(err.Error())
				os.Exit(1)
			}
			removed, err := remove.RemoveAllVers(pkg, pkgConf)
			if err != nil {
				color.Red(err.Error())
				os.Exit(1)
			}
			if removed {
				fmt.Println("Removed", color.CyanString(pkg))
			} else {
				color.HiBlack("%s was not previously installed", pkg)
			}
		}
		fmt.Println("Successfully removed all packages in group ", color.YellowString(group))
	},
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.webman.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	RemoveCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "remove all versions of the packages in group")
}

package remove

import (
	"fmt"
	"os"

	"github.com/candrewlee14/webman/cmd/remove"
	"github.com/candrewlee14/webman/pkgparse"
	"github.com/candrewlee14/webman/utils"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var allFlag bool

var RemoveCmd = &cobra.Command{
	Use:   "remove [group]",
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
		groupConf := pkgparse.ParseGroupConfig(group)
		var pkgsToRemove []string
		if allFlag {
			pkgsToRemove = groupConf.Packages
		} else {
			surveyPrompt := &survey.MultiSelect{
				Message:  "Select packages from group " + color.YellowString(group) + " to " + color.RedString("remove") + ":",
				Options:  groupConf.Packages,
				PageSize: 10,
			}
			err := survey.AskOne(surveyPrompt, &pkgsToRemove)
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}
		}
		if len(pkgsToRemove) == 0 {
			color.HiBlack("No packages selected for removal.")
			os.Exit(0)
		}
		for _, pkg := range pkgsToRemove {
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
		fmt.Printf("All %d selected packages in group %s are uninstalled.\n", len(pkgsToRemove), color.YellowString(group))
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

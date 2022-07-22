package remove

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/candrewlee14/webman/cmd/remove"
	"github.com/candrewlee14/webman/config"
	"github.com/candrewlee14/webman/pkgparse"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
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
				return fmt.Errorf("Prompt failed %v\n", err)
			}
		}
		if len(pkgsToRemove) == 0 {
			color.HiBlack("No packages selected for removal.")
			os.Exit(0)
		}
		for _, pkg := range pkgsToRemove {
			pkgConf, err := pkgparse.ParsePkgConfigLocal(cfg.PkgRepos, pkg)
			if err != nil {
				return err
			}
			removed, err := remove.RemoveAllVers(pkg, pkgConf)
			if err != nil {
				return err
			}
			if removed {
				fmt.Println("Removed", color.CyanString(pkg))
			} else {
				color.HiBlack("%s was not previously installed", pkg)
			}
		}
		fmt.Printf("All %d selected packages in group %s are uninstalled.\n", len(pkgsToRemove), color.YellowString(group))
		return nil
	},
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	RemoveCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "remove all versions of the packages in group")
}

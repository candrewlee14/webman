package remove

import (
	"fmt"
	"os"

	"github.com/candrewlee14/webman/cmd/remove"
	"github.com/candrewlee14/webman/config"
	"github.com/candrewlee14/webman/pkgparse"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Help()
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		group := args[0]
		groupConf, _, err := pkgparse.ParseGroupConfigLocal(cfg.PkgRepos, group)
		if err != nil {
			return err
		}

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
				fmt.Print(pkgConf.RemoveNotes())
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
	RemoveCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "remove all versions of the packages in group")
}

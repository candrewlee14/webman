package add

import (
	"errors"
	"fmt"
	"os"

	"github.com/candrewlee14/webman/cmd/add"
	"github.com/candrewlee14/webman/config"
	"github.com/candrewlee14/webman/multiline"
	"github.com/candrewlee14/webman/pkgparse"
	"github.com/candrewlee14/webman/utils"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	doRefresh bool
	allFlag   bool
)

var AddCmd = &cobra.Command{
	Use:   "add [group]",
	Short: "install a group of packages",
	Long: `

The "group add" subcommand installs a group of packages.
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
			return errors.New("Expected a single package group name")
		}
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
						fmt.Println(err)
					} else {
						color.HiBlue("%s%sRefreshed package recipes!",
							multiline.MoveUp, multiline.ClearLine)
					}
				}
			}
		}
		group := args[0]
		groupConf := pkgparse.ParseGroupConfig(group)

		var pkgsToInstall []string
		if allFlag {
			pkgsToInstall = groupConf.Packages
		} else {
			pkgInfos, err := pkgparse.ParseMultiPkgInfo(groupConf.Packages)
			if err != nil {
				color.Red("failed to parse package info: %v", err)
			}
			infoLines := make([]string, len(pkgInfos))
			for i, pkgInfo := range pkgInfos {
				infoLines[i] = color.CyanString(pkgInfo.Title) + color.HiBlackString(" - ") + pkgInfo.Tagline
			}
			prompt := &survey.MultiSelect{
				Message:  "Select packages from group " + color.YellowString(group) + " to install:",
				Options:  infoLines,
				PageSize: 10,
			}
			var indices []int
			survey.AskOne(prompt, &indices)
			for _, val := range indices {
				pkgsToInstall = append(pkgsToInstall, groupConf.Packages[val])
			}
		}
		if len(pkgsToInstall) == 0 {
			color.HiBlack("No packages selected for installation.")
		} else {
			if !add.InstallAllPkgs(cfg.PkgRepos, pkgsToInstall) {
				color.Magenta("Not all packages installed successfully")
				os.Exit(1)
			}
			color.Green("All %d selected packages from group %s are installed", len(pkgsToInstall), color.YellowString(group))
		}
		return nil
	},
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	AddCmd.Flags().BoolVar(&doRefresh, "refresh", false, "force refresh of package recipes")
	AddCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "add latest versions of all packages in group")
}

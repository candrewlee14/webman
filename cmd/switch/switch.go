package switchcmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/candrewlee14/webman/link"
	"github.com/candrewlee14/webman/pkgparse"
	"github.com/candrewlee14/webman/utils"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// SwitchCmd represents the remove command
var SwitchCmd = &cobra.Command{
	Use:   "switch [pkg]",
	Short: "switch to a specific version of a package",
	Long:  `The "switch" subcommand changes path to a prompt-selected version of a given package.`,
	Example: `webman switch go
webman switch zig
webman switch rg`,
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.Init()
		if len(args) != 1 {
			cmd.Help()
			return nil
		}
		pkg := args[0]
		pkgDir := filepath.Join(utils.WebmanPkgDir, pkg)
		dirEntries, err := os.ReadDir(pkgDir)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("No versions of %s are currently installed.\n", pkg)
			}
			return err
		}

		using, err := pkgparse.CheckUsing(pkg)
		if err != nil {
			return err
		}
		if using != nil {
			fmt.Println("Currently using: ", color.CyanString(*using))
		} else {
			fmt.Printf("Not currently using any %s version\n", color.CyanString(pkg))
		}

		var pkgVersions []string
		for _, entry := range dirEntries {
			if entry.IsDir() {
				pkgVersions = append(pkgVersions, entry.Name())
			}
		}
		pkgConf, err := pkgparse.ParsePkgConfigLocal(pkg, false)
		if err != nil {
			return err
		}
		var pkgVerStem string
		if len(pkgVersions) == 1 {
			pkgVerStem = pkgVersions[0]
			if using != nil && *using == pkgVerStem {
				fmt.Printf("Only one version of %s installed, which is already in use.\n", pkg)
				return nil
			}
		} else {
			surveyPrompt := &survey.Select{
				Message: "Select " + color.CyanString(pkg) + " version to switch to use:",
				Options: pkgVersions,
			}
			err := survey.AskOne(surveyPrompt, &pkgVerStem)
			if err != nil {
				return fmt.Errorf("Prompt failed %v\n", err)
			}
		}
		binPaths, err := pkgConf.GetMyBinPaths()
		if err != nil {
			return err
		}
		_, ver := utils.ParseStem(pkgVerStem)
		madeLinks, err := link.CreateLinks(pkg, ver, binPaths)
		if err != nil {
			return err
		}
		if !madeLinks {
			return fmt.Errorf("Unable to create all links")
		}
		fmt.Printf("Created links for %s\n", pkgVerStem)
		color.Green("Successfully switched, %s now using %s\n", pkg, color.CyanString(pkgVerStem))
		return nil
	},
}

func init() {
	//rootCmd.AddCommand(switchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// removeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

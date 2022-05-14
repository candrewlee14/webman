package switchcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"webman/link"
	"webman/pkgparse"
	"webman/utils"

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
	Run: func(cmd *cobra.Command, args []string) {
		utils.Init()
		if len(args) != 1 {
			cmd.Help()
			os.Exit(0)
		}
		pkg := args[0]
		pkgDir := filepath.Join(utils.WebmanPkgDir, pkg)
		dirEntries, err := os.ReadDir(pkgDir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("No versions of %s are currently installed.\n", color.CyanString(pkg))
				os.Exit(0)
			}
			panic(err)
		}

		using, err := pkgparse.CheckUsing(pkg)
		if err != nil {
			panic(err)
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
			color.Red("%v", err)
			os.Exit(1)
		}
		var pkgVerStem string
		if len(pkgVersions) == 1 {
			pkgVerStem = pkgVersions[0]
			if using != nil && *using == pkgVerStem {
				fmt.Printf("Only one version of %s installed, which is already in use.\n", pkg)
				os.Exit(0)
			}
		} else {
			surveyPrompt := &survey.Select{
				Message: "Select " + color.CyanString(pkg) + " version to switch to use:",
				Options: pkgVersions,
			}
			err := survey.AskOne(surveyPrompt, &pkgVerStem)
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				os.Exit(1)
			}
		}
		binPaths, err := pkgConf.GetMyBinPaths()
		if err != nil {
			fmt.Println(color.RedString("%v", err))
			return
		}
		madeLinks, err := link.CreateLinks(pkg, pkgVerStem, binPaths)
		if err != nil {
			panic(err)
		}
		if !madeLinks {
			panic("Unable to create all links")
		}
		fmt.Printf("Created links for %s\n", pkgVerStem)
		color.Green("Successfully switched, %s now using %s\n", pkg, color.CyanString(pkgVerStem))
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

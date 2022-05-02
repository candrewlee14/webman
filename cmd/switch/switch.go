package switchcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"webman/link"
	cmdutils "webman/pkg/cmd-utils"
	"webman/pkgparse"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// SwitchCmd represents the remove command
var SwitchCmd = &cobra.Command{
	Use:   "switch",
	Short: "switch to a specific version of a package",
	Long:  `The "switch" subcommand changes path to a prompt-selected version of a given package.`,
	Example: `webman switch go
webman switch zig
webman switch rg`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Help()
			os.Exit(0)
		}
		pkg := args[0]
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		webmanDir := filepath.Join(homeDir, ".webman")
		pkgDir := filepath.Join(webmanDir, "pkg", pkg)

		dirEntries, err := os.ReadDir(pkgDir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("No versions of %s are currently installed.\n", color.CyanString(pkg))
				os.Exit(0)
			}
			panic(err)
		}

		using, err := pkgparse.CheckUsing(pkg, webmanDir)
		if err != nil {
			panic(err)
		}
		if using != nil {
			fmt.Println("Currently using: ", color.YellowString(*using))
		} else {
			fmt.Printf("Not currently using any %s version\n", color.CyanString(pkg))
		}

		var pkgVersions []string
		for _, entry := range dirEntries {
			if entry.IsDir() {
				pkgVersions = append(pkgVersions, entry.Name())
			}
		}
		cmdutils.RecipeDir = filepath.Join(webmanDir, "recipes")
		pkgConf, err := pkgparse.ParsePkgConfigLocal(cmdutils.RecipeDir, pkg)
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
			prompt := promptui.Select{
				Label: "Select " + color.CyanString(pkg) + " version to switch to use",
				Items: pkgVersions,
			}
			_, pkgVerStem, err = prompt.Run()
			if err != nil {
				panic(err)
			}
		}
		binPath, err := pkgConf.GetMyBinPath()
		if err != nil {
			fmt.Println(color.RedString("%v", err))
			return
		}
		madeLinks, err := link.CreateLinks(webmanDir, pkg, pkgVerStem, binPath)
		if err != nil {
			panic(err)
		}
		if !madeLinks {
			panic("Unable to create all links")
		}
		fmt.Printf("Created links for %s\n", pkgVerStem)
		color.Green("Successfully switched, %s now using %s\n", pkg, color.YellowString(pkgVerStem))
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

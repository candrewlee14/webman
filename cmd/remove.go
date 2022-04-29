package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"webman/link"
	"webman/multiline"
	"webman/pkgparse"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove a package",
	Long:  `The "remove" subcommand removes a prompt-selected version of a given package.`,
	Example: `webman remove go
webman remove zig
webman remove rg`,
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
		var pkgVerStem string
		if len(pkgVersions) == 1 {
			pkgVerStem = pkgVersions[0]
		} else {

			prompt := promptui.Select{
				Label: "Select " + color.CyanString(pkg) + " version to " + color.RedString("remove"),
				Items: pkgVersions,
			}
			_, pkgVerStem, err = prompt.Run()
		}

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}
		pkgVerDir := filepath.Join(pkgDir, pkgVerStem)
		pkgConf, err := pkgparse.ParsePkgConfig(pkg)
		if err != nil {
			panic(err)
		}
		if using != nil && *using == pkgVerStem {
			binPath, err := pkgConf.GetMyBinPath()
			if err != nil {
				fmt.Println(color.RedString("%v", err))
				return
			}
			_, linkPaths, err := link.GetBinPathsAndLinkPaths(webmanDir, pkg, pkgVerStem, binPath)
			if err != nil {
				panic(err)
			}
			fmt.Println("Removing links ...")
			for _, linkPath := range linkPaths {
				if runtime.GOOS == "windows" {
					linkPath = linkPath + ".bat"
				}
				err := os.Remove(linkPath)
				if err != nil {
					panic(err)
				}
			}
			fmt.Printf("%s%sRemoved links!\n", multiline.MoveUp, multiline.ClearLine)
			if err = pkgparse.RemoveUsing(pkg, webmanDir); err != nil {
				panic(err)
			}
		}
		// Remove directory
		fmt.Printf("Removing %s ...\n", pkgVerStem)
		// if this is the only version of this package installed, remove this pkg's whole dir
		if len(pkgVersions) == 1 {
			if err := os.RemoveAll(pkgDir); err != nil {
				panic(err)
			}
		} else { // otherwise just remove the pkg version's dir
			if err := os.RemoveAll(pkgVerDir); err != nil {
				panic(err)
			}
		}
		fmt.Printf("%s%sRemoved %s!\n", multiline.MoveUp, multiline.ClearLine, pkgVerStem)
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// removeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

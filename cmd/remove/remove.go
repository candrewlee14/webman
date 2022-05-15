package remove

import (
	"fmt"
	"os"
	"path/filepath"
	"webman/link"
	"webman/multiline"
	"webman/pkgparse"
	"webman/utils"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// RemoveCmd represents the remove command
var RemoveCmd = &cobra.Command{
	Use:   "remove [pkg]",
	Short: "remove a package",
	Long:  `The "remove" subcommand removes a prompt-selected version of a given package.`,
	Example: `webman remove go
webman remove zig
webman remove rg`,
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
		var pkgVerStems []string
		if len(pkgVersions) == 1 {
			pkgVerStems = append(pkgVerStems, pkgVersions[0])
		} else {
			surveyPrompt := &survey.MultiSelect{
				Message:  "Select " + color.CyanString(pkg) + " version to " + color.RedString("remove") + ":",
				Options:  pkgVersions,
				PageSize: 10,
			}
			err := survey.AskOne(surveyPrompt, &pkgVerStems)
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}
		}
		if len(pkgVerStems) == 0 {
			color.HiBlack("No packages selected for removal.")
			os.Exit(0)
		}
		pkgConf, err := pkgparse.ParsePkgConfigLocal(pkg, false)
		if err != nil {
			panic(err)
		}
		// if we are installing all versions, remove the whole directory
		if len(pkgVerStems) == len(pkgVersions) {
			if _, err := RemoveAllVers(pkg, pkgConf); err != nil {
				panic(err)
			}
		} else {
			for _, pkgVerStem := range pkgVerStems {
				RemovePkgVer(pkgVerStem, using, pkg, pkgConf)
			}
		}
		fmt.Printf("All %d selected packages are uninstalled.\n", len(pkgVerStems))
	},
}

// Uninstalls the binaries for a package (if they are installed)
func UninstallBins(pkg string, pkgConf *pkgparse.PkgConfig) error {
	using, err := pkgparse.CheckUsing(pkg)
	if err != nil {
		return err
	}
	if using == nil {
		return nil
	}
	pkgVerStem := *using
	binPaths, err := pkgConf.GetMyBinPaths()
	if err != nil {
		return err
	}
	_, ver := utils.ParseStem(pkgVerStem)
	_, linkPaths, err := link.GetBinPathsAndLinkPaths(pkg, ver, binPaths)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Removing %s links ...\n", color.CyanString(pkg))
	for _, linkPath := range linkPaths {
		if utils.GOOS == "windows" {
			linkPath = linkPath + ".bat"
		}
		err := os.Remove(linkPath)
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("%s%sRemoved %s links!\n", multiline.MoveUp, multiline.ClearLine, color.CyanString(pkg))
	if err = pkgparse.RemoveUsing(pkg); err != nil {
		return err
	}
	return nil
}

func RemovePkgVer(pkgVerStem string, using *string, pkg string, pkgConf *pkgparse.PkgConfig) {
	// if the selected pkgVerStem is being used, uninstall bins
	if using != nil && *using == pkgVerStem {
		if err := UninstallBins(pkg, pkgConf); err != nil {
			color.Red("Error uninstalling binaries: %v", err)
			os.Exit(1)
		}
	}
	fmt.Printf("Removing %s ...\n", pkgVerStem)
	if err := os.RemoveAll(filepath.Join(utils.WebmanPkgDir, pkg, pkgVerStem)); err != nil {
		panic(err)
	}
	fmt.Printf("%s%sRemoved %s!\n", multiline.MoveUp, multiline.ClearLine, pkgVerStem)
}

func RemoveAllVers(pkg string, pkgConf *pkgparse.PkgConfig) (bool, error) {
	if err := UninstallBins(pkg, pkgConf); err != nil {
		return false, err
	}
	pkgDir := filepath.Join(utils.WebmanPkgDir, pkg)
	if _, err := os.Stat(pkgDir); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if err := os.RemoveAll(pkgDir); err != nil {
		return false, err
	}
	return true, nil
}
func GetPkgVerStems(pkg string) error {
	return nil
}

func init() {
	//rootCmd.AddCommand(removeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// removeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

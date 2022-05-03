package remove

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"webman/link"
	"webman/multiline"
	"webman/pkgparse"
	"webman/utils"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// RemoveCmd represents the remove command
var RemoveCmd = &cobra.Command{
	Use:   "remove",
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
		pkgConf, err := pkgparse.ParsePkgConfigLocal(pkg, false)
		if err != nil {
			panic(err)
		}
		// if the selected pkgVerStem is being used, uninstall bins
		if using != nil && *using == pkgVerStem {
			if err = UninstallBins(pkg, pkgConf); err != nil {
				color.Red("Error uninstalling binaries: %v", err)
				os.Exit(1)
			}
		}
		fmt.Printf("Removing %s ...\n", pkgVerStem)
		// if this is the only version of this package installed, remove this pkg's whole dir
		if len(pkgVersions) == 1 {
			if _, err := RemoveAllVers(pkg, pkgConf); err != nil {
				panic(err)
			}
		} else { // otherwise just remove the pkg version's dir
			if err := os.RemoveAll(filepath.Join(utils.WebmanPkgDir, pkg, pkgVerStem)); err != nil {
				panic(err)
			}
		}
		fmt.Printf("%s%sRemoved %s!\n", multiline.MoveUp, multiline.ClearLine, pkgVerStem)
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
	binPath, err := pkgConf.GetMyBinPath()
	if err != nil {
		return err
	}
	_, linkPaths, err := link.GetBinPathsAndLinkPaths(pkg, pkgVerStem, binPath)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Removing %s links ...\n", color.CyanString(pkg))
	for _, linkPath := range linkPaths {
		if runtime.GOOS == "windows" {
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

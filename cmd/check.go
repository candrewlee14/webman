package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"webman/pkgparse"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "check a directory of recipes",
	Long: `
The "check" subcommand checks that all recipes in a directory are valid.`,
	Example: `webman check ~/repos/webman-pkgs/pkg
webman check ~/repos/webman-pkgs/pkg/go.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Help()
			os.Exit(0)
		}
		recipeDirOrFile, err := filepath.Abs(args[0])
		if err != nil {
			panic(err)
		}
		entries, err := os.ReadDir(filepath.Join(recipeDirOrFile, "pkgs"))
		if err != nil {
			panic(err)
		} else {
			var wg sync.WaitGroup
			success := true
			wg.Add(len(entries))
			for _, recipe := range entries {
				recipe := recipe

				go func() {
					recipeName := recipe.Name()
					err := CheckPkgConfig(recipeDirOrFile, recipeName)
					if err != nil {
						color.Red("%s: %s", color.YellowString(recipeName), color.RedString("%v", err))
						success = false
					}
					wg.Done()
				}()
			}
			wg.Wait()
			if !success {
				color.Magenta("Not all packages are valid!")
				os.Exit(1)
			}
			color.Green("All packages are valid!")
		}
	},
}

func CheckPkgConfig(recipeDir string, pkgFileName string) error {
	pkg := strings.ReplaceAll(pkgFileName, ".yaml", "")
	pkgConf, err := pkgparse.ParsePkgConfigLocal(recipeDir, pkg)
	if err != nil {
		return err
	}
	if len(pkgConf.FilenameFormat) == 0 {
		return fmt.Errorf("filename_format field empty")
	}
	if len(pkgConf.BaseDownloadUrl) == 0 {
		return fmt.Errorf("base_download_url field empty")
	}
	if len(pkgConf.LatestStrategy) == 0 {
		return fmt.Errorf("latest_strategy field empty")
	}
	switch pkgConf.LatestStrategy {
	case "github-release":
		if len(pkgConf.GitUser) == 0 {
			return fmt.Errorf("missing git_user because github-release latest strategy")
		}
		if len(pkgConf.GitRepo) == 0 {
			return fmt.Errorf("missing git_repo because github-release latest strategy")
		}
	case "arch-linux-community":
		if len(pkgConf.ArchLinuxPkgName) == 0 {
			return fmt.Errorf("missing arch_linux_pkg_name because arch-linux-community latest strategy")
		}
	default:
		return fmt.Errorf("invalid latest strategy")
	}
	return nil
}

func init() {
	rootCmd.AddCommand(checkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// removeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
package check

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/candrewlee14/webman/pkgparse"
	"github.com/candrewlee14/webman/schema"
	"github.com/candrewlee14/webman/utils"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// CheckCmd represents the remove command
var CheckCmd = &cobra.Command{
	Use:   "check [recipe-dir]",
	Short: "check a directory of recipes",
	Long: `
The "check" subcommand checks that all recipes in a directory are valid.`,
	Example: `webman check ~/repos/webman-pkgs/`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Help()
		}
		recipeDir, err := filepath.Abs(args[0])
		utils.WebmanRecipeDir = recipeDir
		if err != nil {
			return err
		}
		entries, err := os.ReadDir(filepath.Join(recipeDir, "pkgs"))
		if err != nil {
			return err
		} else {
			var wg sync.WaitGroup
			success := true
			wg.Add(len(entries))
			for _, recipe := range entries {
				recipe := recipe

				go func() {
					recipeName := recipe.Name()
					pkg := strings.ReplaceAll(recipeName, utils.PkgRecipeExt, "")
					err := CheckPkgConfig(pkg)
					if err != nil {
						var lintErr schema.ResultErrors
						if errors.As(err, &lintErr) {
							for _, le := range lintErr {
								color.Red("%s: %s", color.YellowString(recipeName), color.RedString("%s: %s", le.Field(), le.Description()))
							}
						} else {
							color.Red("%s: %s", color.YellowString(recipeName), color.RedString("%v", err))
						}
						success = false
					}
					wg.Done()
				}()
			}
			wg.Wait()
			if !success {
				return fmt.Errorf("Not all packages are valid!")
			}
			color.Green("All packages are valid!")

			groupsPath := filepath.Join(recipeDir, "groups")
			groups, err := os.ReadDir(groupsPath)
			if err != nil {
				return err
			}
			var wg2 sync.WaitGroup
			success = true
			wg2.Add(len(groups))
			for _, groupEntry := range groups {
				groupEntry := groupEntry
				go func() {
					recipeName := groupEntry.Name()
					group := strings.ReplaceAll(recipeName, utils.GroupRecipeExt, "")
					if err := CheckGroup(filepath.Join(groupsPath, recipeName)); err != nil {
						color.Red("%s: %s", color.MagentaString(group), color.RedString("%v", err))
						success = false
					}
					wg2.Done()
				}()
			}
			wg2.Wait()
			if !success {
				return fmt.Errorf("Not all groups are valid!")
			}
			color.Green("All groups are valid!")
			return nil
		}
	},
}

func CheckGroup(path string) error {
	group := filepath.Base(path)
	fi, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fi.Close()
	groupConf, err := pkgparse.ParseGroupConfig(fi, group)
	if err != nil {
		return fmt.Errorf("no group file found for %s: %v", group, err)
	}
	for _, pkg := range groupConf.Packages {
		if err := CheckPkgConfig(pkg); err != nil {
			return err
		}
	}
	return nil
}

func CheckPkgConfig(pkg string) error {
	pkgConfPath := filepath.Join(utils.WebmanRecipeDir, "pkgs", pkg+utils.PkgRecipeExt)
	fi, err := os.Open(pkgConfPath)
	if err != nil {
		return err
	}
	defer fi.Close()
	return schema.LintRecipe(fi)
}

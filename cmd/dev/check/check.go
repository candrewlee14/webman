package check

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

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
			return nil
		}
	},
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

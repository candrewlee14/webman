package check

import (
	"errors"
	"github.com/candrewlee14/webman/schema"
	"os"
	"path/filepath"
	"strings"
	"sync"

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
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Help()
			os.Exit(0)
		}
		recipeDir, err := filepath.Abs(args[0])
		utils.WebmanRecipeDir = recipeDir
		if err != nil {
			panic(err)
		}
		entries, err := os.ReadDir(filepath.Join(recipeDir, "pkgs"))
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
				color.Magenta("Not all packages are valid!")
				os.Exit(1)
			}
			color.Green("All packages are valid!")
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
	return schema.Lint(fi)
}

func init() {

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// removeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

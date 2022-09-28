package check

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/candrewlee14/webman/config"
	"github.com/candrewlee14/webman/utils"

	"github.com/fatih/color"
)

// NestedRecipe v0.8.0 -> v0.9.0 introduced configurable repos, which moved the recipes dir one level deeper
var NestedRecipe = Check{
	Name: "Nested Recipes",
	Func: func(cfg *config.Config, fix bool) error {
		_, pkgsErr := os.Stat(filepath.Join(utils.WebmanRecipeDir, "pkgs"))
		_, groupsErr := os.Stat(filepath.Join(utils.WebmanRecipeDir, "groups"))
		if errors.Is(pkgsErr, fs.ErrNotExist) && errors.Is(groupsErr, fs.ErrNotExist) {
			color.HiGreen("no un-nested recipes detected")
			return nil
		}

		if !fix {
			color.HiRed("detected un-nested recipes, please clean up %q", utils.WebmanRecipeDir)
			return nil
		}

		// As of https://github.com/candrewlee14/webman-pkgs/tree/9eda7908a9f25398c1c693e4ec7dc7a727eddf71
		remove := []string{".github", "groups", "pkgs", ".mega-linter.yml", ".pre-commit-config.yaml", "LICENSE", "README.md", "full-bintest.sh", "refresh.yaml"}
		for _, r := range remove {
			if err := os.RemoveAll(filepath.Join(utils.WebmanRecipeDir, r)); err != nil {
				color.HiRed("could not remove %q: %v", r, err)
			}
		}

		color.HiGreen("successfully cleaned un-nested recipes")
		return nil
	},
}

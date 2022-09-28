package check

import (
	"os"
	"path/filepath"

	"github.com/candrewlee14/webman/config"
	"github.com/candrewlee14/webman/link"
	"github.com/candrewlee14/webman/pkgparse"
	"github.com/candrewlee14/webman/utils"

	"github.com/fatih/color"
)

// Windows switched from batch files to symlinks
var WindowsSymlink = Check{
	Name: "Windows Symlink",
	Func: func(cfg *config.Config, fix bool) error {
		if utils.GOOS != "windows" {
			color.HiGreen("current OS isn't windows; skipping")
			return nil
		}

		installed, err := os.ReadDir(utils.WebmanPkgDir)
		if err != nil {
			return err
		}

		for _, i := range installed {
			pkgConfig, err := pkgparse.ParsePkgConfigLocal(cfg.PkgRepos, i.Name())
			if err != nil {
				return err
			}

			binPaths, err := pkgConfig.GetMyBinPaths()
			if err != nil {
				return err
			}
			renames, err := pkgConfig.GetRenames()
			if err != nil {
				return err
			}
			ver, err := pkgparse.CheckUsing(i.Name())
			if err != nil {
				return err
			}

			if !fix {
				_, err := os.Stat(filepath.Join(utils.WebmanBinDir, i.Name()+".exe"))
				if err != nil {
					color.HiRed("could not find symlink for %q: %v", i.Name(), err)
				}
				color.HiGreen("symlink found for %s", i.Name())
				continue
			}

			if _, err := link.CreateLinks(i.Name(), *ver, binPaths, renames); err != nil {
				color.HiRed("could not create symlink(s) for %q: %v", i.Name(), err)
			}
		}

		if fix {
			bats, err := filepath.Glob(filepath.Join(utils.WebmanBinDir, "*.bat"))
			if err != nil {
				return err
			}
			for _, bat := range bats {
				if err := os.Remove(bat); err != nil {
					color.HiRed("could not remove batch file %q: %v", bat, err)
				}
			}
		}

		return nil
	},
}

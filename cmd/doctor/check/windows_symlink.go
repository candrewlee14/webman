package check

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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
			using, err := pkgparse.CheckUsing(i.Name())
			if err != nil {
				return err
			}
			parts := strings.Split(*using, "-")
			ver := parts[len(parts)-1]

			_, err = os.Lstat(filepath.Join(utils.WebmanBinDir, i.Name()+".exe"))
			if err == nil {
				continue
			}
			if err != nil {
				if !errors.Is(err, fs.ErrNotExist) {
					color.HiRed("could not lstat %q: %v", i.Name(), err)
					continue
				}
			}

			if !fix {
				color.HiRed("no symlink(s) found for %q", i.Name())
				continue
			}

			color.HiGreen("creating symlink(s) for %s", i.Name())
			if _, err := link.CreateLinks(i.Name(), ver, binPaths, renames); err != nil {
				color.HiRed("could not create symlink(s) for %q: %v", i.Name(), err)
			}
		}

		bats, err := filepath.Glob(filepath.Join(utils.WebmanBinDir, "*.bat"))
		if err != nil {
			return err
		}

		if len(bats) == 0 {
			color.HiGreen("no batch files found")
			return nil
		}

		if !fix {
			color.HiRed("found %d batch file(s)", len(bats))
			return nil
		}

		color.HiGreen("removing batch file(s)")
		for _, bat := range bats {
			if err := os.Remove(bat); err != nil {
				color.HiRed("could not remove batch file %q: %v", bat, err)
			}
		}

		return nil
	},
}

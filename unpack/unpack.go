package unpack

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mholt/archiver/v3"

	"github.com/candrewlee14/webman/utils"
)

func Unpack(src string, pkg string, stem string, hasRoot bool) error {
	pkgDir := filepath.Join(utils.WebmanPkgDir, pkg)
	err := os.MkdirAll(pkgDir, 0755)
	if err != nil {
		return fmt.Errorf("unable to create dir %q: %v", pkgDir, err)
	}
	pkgDest := filepath.Join(pkgDir, stem)
	if hasRoot {
		tmpPkgDir := filepath.Join(utils.WebmanTmpDir, pkg)
		if err := os.MkdirAll(tmpPkgDir, 0755); err != nil {
			return fmt.Errorf("unable to create dir %q: %v", tmpPkgDir, err)
		}
		if err := archiver.Unarchive(src, tmpPkgDir); err != nil {
			return fmt.Errorf("failed to extract file: %v", err)
		}
		f, err := os.Open(tmpPkgDir)
		if err != nil {
			return fmt.Errorf("unable to open dir %q: %v", tmpPkgDir, err)
		}
		dir, err := f.ReadDir(1)
		if err != nil {
			return fmt.Errorf("unable to read dir %q: %v", tmpPkgDir, err)
		}
		extractFolder := filepath.Join(tmpPkgDir, dir[0].Name())
		if err = os.Rename(extractFolder, pkgDest); err != nil {
			return fmt.Errorf("unable to move %q to %q: %v", extractFolder, pkgDest, err)
		}
	}
	if err := os.MkdirAll(pkgDest, 0777); err != nil {
		return fmt.Errorf("unable to create pkg destination dir %q: %v", pkgDest, err)
	}
	if err := archiver.Unarchive(src, pkgDest); err != nil {
		return fmt.Errorf("failed to extract file: %v", err)
	}
	return nil
}

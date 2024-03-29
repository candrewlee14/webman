package unpack

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/candrewlee14/webman/utils"

	"github.com/mholt/archiver/v3"
)

func Unpack(src string, pkg string, stem string, hasRoot bool) error {
	pkgDir := filepath.Join(utils.WebmanPkgDir, pkg)
	err := os.MkdirAll(pkgDir, 0o755)
	if err != nil {
		return fmt.Errorf("unable to create dir %q: %v", pkgDir, err)
	}
	pkgDest := filepath.Join(pkgDir, stem)
	if hasRoot {
		tmpPkgDir := filepath.Join(utils.WebmanTmpDir, pkg)
		if err := os.MkdirAll(tmpPkgDir, 0o755); err != nil {
			return fmt.Errorf("unable to create dir %q: %v", tmpPkgDir, err)
		}
		if err := unpack(pkg, src, tmpPkgDir); err != nil {
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
	} else {
		if err := os.MkdirAll(pkgDest, 0o777); err != nil {
			return fmt.Errorf("unable to create pkg destination dir %q: %v", pkgDest, err)
		}
		if err := unpack(pkg, src, pkgDest); err != nil {
			return fmt.Errorf("failed to extract file: %v", err)
		}
	}
	return nil
}

func unpack(pkg, src, dest string) error {
	uaIface, err := archiver.ByExtension(src)
	if err != nil {
		return err
	}
	u, ok := uaIface.(archiver.Unarchiver)
	if !ok {
		d, ok := uaIface.(archiver.Decompressor)
		if !ok {
			return fmt.Errorf("format specified by source filename is not an archive or compression format: %s (%T)", src, uaIface)
		}
		var binExt string
		if utils.GOOS == "windows" {
			binExt = ".exe"
		}
		dest = filepath.Join(dest, pkg+binExt)
		c := archiver.FileCompressor{Decompressor: d}
		if err := c.DecompressFile(src, dest); err != nil {
			return err
		}
		return os.Chmod(dest, 0o755)
	}
	return u.Unarchive(src, dest)
}

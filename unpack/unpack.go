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
	}
	if err := os.MkdirAll(pkgDest, 0777); err != nil {
		return fmt.Errorf("unable to create pkg destination dir %q: %v", pkgDest, err)
	}
	if err := unpack(pkg, src, pkgDest); err != nil {
		return fmt.Errorf("failed to extract file: %v", err)
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
		dest = filepath.Join(dest, pkg)
		c := archiver.FileCompressor{Decompressor: d}
		if err := c.DecompressFile(src, dest); err != nil {
			return err
		}
		return os.Chmod(dest, 0755)
	}
	return u.Unarchive(src, dest)
}

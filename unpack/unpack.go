package unpack

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"webman/utils"

	"github.com/ulikunitz/xz"
)

type unpackExt string

const (
	extTarGz unpackExt = "tar.gz"
	extTarXz unpackExt = "tar.xz"
	extZip   unpackExt = "zip"
	extGz    unpackExt = "gz"
)

type UnpackFn func(src string, dir string) error

var unpackMap = map[unpackExt]UnpackFn{
	extTarGz: UntarGz,
	extTarXz: UntarXz,
	extGz:    UnGz,
	extZip:   Unzip,
}

func Unpack(src string, pkg string, stem string, ext string, hasRoot bool) error {
	unpackFn, exists := unpackMap[unpackExt(ext)]
	if !exists {
		return fmt.Errorf("no unpack function for extension: %q", ext)
	}
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
		if err = unpackFn(src, tmpPkgDir); err != nil {
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
		if err := os.MkdirAll(pkgDest, 0777); err != nil {
			return fmt.Errorf("unable to create pkg destination dir %q: %v", pkgDest, err)
		}
		// if this is a gzipped binary
		if ext == "gz" {
			pkgDest = filepath.Join(pkgDest, pkg)
		}
		if err = unpackFn(src, pkgDest); err != nil {
			return fmt.Errorf("failed to extract file: %v", err)
		}
	}
	return nil
}

func UntarXz(src string, dir string) error {
	// Open compress file
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	// Add xz support
	uncompressedStream, err := xz.NewReader(file)
	if err != nil {
		return err
	}
	return Untar(uncompressedStream, dir)
}

func UntarGz(src string, dir string) error {
	// Open compress file
	// Open compress file
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	// Add gzip support
	uncompressedStream, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer uncompressedStream.Close()
	return Untar(uncompressedStream, dir)
}

func UnGz(src string, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	// Add gzip support
	uncompressedStream, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer uncompressedStream.Close()

	destFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer destFile.Close()
	if _, err := io.Copy(destFile, uncompressedStream); err != nil {
		return err
	}
	return nil
}

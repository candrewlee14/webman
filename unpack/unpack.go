package unpack

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"webman/utils"
)

type UnpackFn func(src string, dir string) error

var unpackMap = map[string]UnpackFn{
	"tar.gz": untarExec,
	"tar.xz": untarExec,
	"zip":    Unzip,
}

func Unpack(src string, pkg string, stem string, ext string, hasRoot bool) error {
	unpackFn, exists := unpackMap[ext]
	if !exists {
		return fmt.Errorf("no unpack function for extension: %q", ext)
	}
	pkgDir := filepath.Join(utils.WebmanPkgDir, pkg)
	err := os.MkdirAll(pkgDir, 0777)
	if err != nil {
		return fmt.Errorf("unable to create dir %q: %v", pkgDir, err)
	}
	pkgDest := filepath.Join(pkgDir, stem)
	if hasRoot {
		tmpPkgDir := filepath.Join(utils.WebmanTmpDir, pkg)
		if err := os.MkdirAll(tmpPkgDir, 0777); err != nil {
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
		if err = unpackFn(src, pkgDest); err != nil {
			return fmt.Errorf("failed to extract file: %v", err)
		}
	}
	return nil
}

func untarExec(src string, dir string) error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("windows doesn't have support for tarballs")
	}
	cmd := exec.Command("tar", "-xf", src, "--directory="+dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Unzip(src string, dir string) error {
	archive, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("unable to unzip: %v", err)
	}
	defer archive.Close()

	for _, f := range archive.File {
		fileName := f.Name

		filePath := filepath.Join(dir, fileName)

		if !strings.HasPrefix(filePath, filepath.Clean(dir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path")
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		fileInArchive, err := f.Open()
		if err != nil {
			return err
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return err
		}

		dstFile.Close()
		fileInArchive.Close()
	}
	return nil
}

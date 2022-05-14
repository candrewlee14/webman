package unpack

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ulikunitz/xz"
)

func Untar(src string, dest string) error {
	// if tar program doesn't exist, default to Go native (unstable)
	if _, err := exec.LookPath("tar"); err != nil {
		switch filepath.Ext(src) {
		case ".tar.gz":
			return UntarGz(src, dest)
		case ".tar.xz":
			return UntarXz(src, dest)
		}
	}
	return UntarExec(src, dest)
}

func UntarGo(uncompressedStream io.Reader, dest string) error {
	// Read content file
	archive := tar.NewReader(uncompressedStream)
	for {
		header, err := archive.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		fileName := header.Name

		filePath := filepath.Join(dest, fileName)

		if !strings.HasPrefix(filePath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path")
		}
		if header.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, header.FileInfo().Mode())
		if err != nil {
			return err
		}

		if _, err := io.Copy(dstFile, archive); err != nil {
			return err
		}

		dstFile.Close()
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
	return UntarGo(uncompressedStream, dir)
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
	return UntarGo(uncompressedStream, dir)
}

func UntarExec(src string, dir string) error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("windows doesn't have support for tarballs")
	}
	cmd := exec.Command("tar", "-xf", src, "--directory="+dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

package unpack

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Untargz(src string, dir string) error {
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

	// Read content file
	archive := tar.NewReader(uncompressedStream)

	var infinityLoop = true
	for infinityLoop {
		header, err := archive.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		fileName := header.Name

		filePath := filepath.Join(dir, fileName)

		if !strings.HasPrefix(filePath, filepath.Clean(dir)+string(os.PathSeparator)) {
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

package unpack

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Untar(uncompressedStream io.Reader, dest string) error {
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

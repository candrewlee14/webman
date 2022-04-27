package unpack

import (
    "archive/zip"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
)


func unzip(src string, dir string) error {
    archive, err := zip.OpenReader(src)
    if err != nil {
        panic(err)
    }
    defer archive.Close()

    for _, f := range archive.File {
        filePath := filepath.Join(dir, f.Name)
        fmt.Println("unzipping file ", filePath)

        if !strings.HasPrefix(filePath, filepath.Clean(dir)+string(os.PathSeparator)) {
            return fmt.Errorf("invalid file path")
        }
        if f.FileInfo().IsDir() {
            fmt.Println("creating directory...")
            os.MkdirAll(filePath, os.ModePerm)
            continue
        }

        if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
            panic(err)
        }

        dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
        if err != nil {
            panic(err)
        }

        fileInArchive, err := f.Open()
        if err != nil {
            panic(err)
        }

        if _, err := io.Copy(dstFile, fileInArchive); err != nil {
            panic(err)
        }

        dstFile.Close()
        fileInArchive.Close()
    }
    return nil
}


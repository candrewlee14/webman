package unpack

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func untarExec(src string, dir string) error {
    if runtime.GOOS == "windows" {
        panic("Windows doesn't have support for tarballs")
    }
    cmd := exec.Command("tar", "-xf", src, "--directory="+dir)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

type UnpackFn func(src string, dir string) error
var unpackMap = map[string]UnpackFn {
    "tar.gz": untarExec,
    "tar.xz": untarExec,
    "zip": unzip,
}

func Unpack(src string, dir string, ext string) error {
    fmt.Println("Unpacking", src);
    unpackFn, exists := unpackMap[ext]
    err := os.MkdirAll(dir, 0777)
    if err != nil {
        return fmt.Errorf("Unable to create dir %q: %v", dir, err)
    }
    if !exists {
        return fmt.Errorf("No unpack function for extension: %q", ext)
    }
    return unpackFn(src, dir)
}

package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var WebmanDir string
var WebmanPkgDir string
var WebmanBinDir string
var WebmanRecipeDir string
var WebmanTmpDir string

func Init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	WebmanDir = filepath.Join(homeDir, "/.webman")
	WebmanPkgDir = filepath.Join(WebmanDir, "/pkg")
	WebmanBinDir = filepath.Join(WebmanDir, "/bin")
	WebmanRecipeDir = filepath.Join(WebmanDir, "/recipes")
	WebmanTmpDir = filepath.Join(WebmanDir, "/tmp")

	if err = os.MkdirAll(WebmanBinDir, os.ModePerm); err != nil {
		panic(err)
	}
	if err = os.MkdirAll(WebmanPkgDir, os.ModePerm); err != nil {
		panic(err)
	}
	if err = os.MkdirAll(WebmanTmpDir, os.ModePerm); err != nil {
		panic(err)
	}
}

func ParsePkgVer(arg string) (string, string, error) {
	parts := strings.Split(arg, "@")
	var pkg string
	var ver string
	if len(parts) == 1 {
		pkg = parts[0]
	} else if len(parts) == 2 {
		pkg = parts[0]
		ver = parts[1]
	} else {
		return "", "", fmt.Errorf("packages should be in format 'pkg' or 'pkg@version'")
	}
	return pkg, ver, nil
}

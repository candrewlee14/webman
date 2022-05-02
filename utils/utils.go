package utils

import (
	"os"
	"path/filepath"
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

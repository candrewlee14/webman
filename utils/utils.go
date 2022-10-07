package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/candrewlee14/webman/multiline"
	"github.com/candrewlee14/webman/ui"

	"github.com/fatih/color"
)

var (
	WebmanDir       string
	WebmanConfig    string
	WebmanPkgDir    string
	WebmanBinDir    string
	WebmanRecipeDir string
	WebmanTmpDir    string
	RecipeDirFlag   string
	GOOS            string
	GOARCH          string
	PkgRecipeExt    = ".webman-pkg.yml"
	GroupRecipeExt  = ".webman-group.yml"
	UsingFileName   = "using.yaml"
)

func Init(homeDir string) {
	WebmanDir = filepath.Join(homeDir, ".webman")
	WebmanConfig = filepath.Join(WebmanDir, "config.yaml")
	WebmanPkgDir = filepath.Join(WebmanDir, "pkg")
	WebmanBinDir = filepath.Join(WebmanDir, "bin")
	WebmanRecipeDir = filepath.Join(WebmanDir, "recipes")
	WebmanTmpDir = filepath.Join(WebmanDir, "tmp")
	GOOS = runtime.GOOS
	GOARCH = runtime.GOARCH

	if err := os.MkdirAll(WebmanBinDir, os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(WebmanPkgDir, os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(WebmanTmpDir, os.ModePerm); err != nil {
		panic(err)
	}
	if RecipeDirFlag != "" {
		recipeDir, err := filepath.Abs(RecipeDirFlag)
		if err != nil {
			color.Red("Failed converting local package directory to absolute path: %v", err)
			os.Exit(1)
		}
		color.Magenta("Using local recipe directory: %s", color.HiBlackString(recipeDir))
		WebmanRecipeDir = recipeDir
	}
	if err := os.MkdirAll(WebmanRecipeDir, os.ModePerm); err != nil {
		panic(err)
	}
	if !ui.AreAnsiCodesEnabled() {
		multiline.ClearLine = []byte{}
		multiline.MoveDown = []byte{}
		multiline.MoveUp = []byte{}
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

func CreateStem(pkg string, ver string) string {
	return fmt.Sprintf("%s-%s", pkg, ver)
}

func ParseStem(pkgVerStem string) (string, string) {
	pkg, ver, _ := strings.Cut(pkgVerStem, "-")
	return pkg, ver
}

// InstalledPackages returns a list of currently installed packages, as per the webman pkgs directory
func InstalledPackages() []string {
	var pkgs []string
	entries, err := os.ReadDir(WebmanPkgDir)
	if err != nil {
		return nil
	}
	for _, entry := range entries {
		pkgs = append(pkgs, entry.Name())
	}
	return pkgs
}

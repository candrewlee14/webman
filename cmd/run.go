package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"webman/pkgparse"

	"github.com/fatih/color"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run installed packages",
	Long: `
The "run" subcommand run packages.`,
	Example: `webman run go
webman run bat
webman add go@18.0.0`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		webmanDir := filepath.Join(homeDir, "/.webman")
		webmanPkgDir := filepath.Join(webmanDir, "/pkg")
		webmanBinDir := filepath.Join(webmanDir, "/bin")
		recipeDir = filepath.Join(webmanDir, "recipes")
		runPackage(args, webmanPkgDir, webmanBinDir, webmanDir)

	},
	DisableFlagParsing: true,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runPackage(args []string, webmanPkgDir, webmanBinDir, webmanDir string) {
	parts := strings.Split(args[0], "@")
	var pkg string
	var ver string
	var pkgRunFolder string
	var pkgBinFile string
	var argsApp []string

	// Version information
	if len(parts) == 1 {
		pkg = parts[0]
	} else if len(parts) == 2 {
		pkg = parts[0]
		ver = parts[1]
	} else {
		exitPrint(1, color.RedString("Packages should be in format 'pkg' or 'pkg@version'"))
	}

	// Add args for pkg
	if len(args) > 1 {
		argsApp = args[1:]
	}

	pkgConf, err := pkgparse.ParsePkgConfigLocal(recipeDir, pkg)
	if err != nil {
		exitPrint(1, color.RedString(err.Error()))
	}

	bin, err := pkgConf.GetMyBinPath()
	if err != nil {
		exitPrint(1, color.RedString(err.Error()))
	}

	// Is custom version
	if ver != "" {
		packageFolderName := fmt.Sprintf("%s-%s", pkg, ver)
		pkgRunFolder = path.Join(webmanPkgDir, pkg, packageFolderName)
	}
	// Default version
	if ver == "" {
		usingVersion, err := pkgparse.CheckUsing(pkg, webmanDir)
		if err != nil {
			panic(err)
		}
		if usingVersion == nil {
			exitPrint(0, fmt.Sprintf("Not currently using any %s version\n", color.CyanString(pkg)))
		}
		pkgRunFolder = path.Join(webmanPkgDir, pkg, *usingVersion)
	}
	pkgBinFile = path.Join(pkgRunFolder, bin)

	// Is folder, pkgBinFile
	pkgBinFileInfo, err := os.Stat(pkgBinFile)
	if err != nil {
		if os.IsNotExist(err) {
			IsNotExist(ver)
		}
		exitPrint(1, color.RedString(err.Error()))
	}
	if pkgBinFileInfo.IsDir() {
		pkgBinFile = path.Join(pkgBinFile, pkg)
	}

	appCmd := exec.Command(pkgBinFile, argsApp...)
	appCmd.Stderr = os.Stderr
	appCmd.Stdout = os.Stdout
	appCmd.Stdin = os.Stdin
	appCmd.Env = os.Environ()

	// Start package
	if err := appCmd.Run(); err != nil {
		if os.IsNotExist(err) {
			IsNotExist(ver)
		}
		exitPrint(1, color.RedString(err.Error()))
	}
}

func IsNotExist(pkg string) {
	fmt.Printf("No versions of %s are currently installed.\n", color.CyanString(pkg))
	os.Exit(0)
}

func exitPrint(code int, text string) {
	fmt.Println(text)
	os.Exit(code)
}

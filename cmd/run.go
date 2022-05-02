package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
The "run" subcommand runs installed package binary with the name of the package.
Pass a double dash before commands you want to forward to the binary.`,
	Example: `webman run go
webman run bat -- [FILE]
webman run go@18.0.0
webman run node@17.0.0 -- --version
webman run node@17.0.0 --select-bin npm -- --version`,
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
}

var selectBin string

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&selectBin, "select-bin", "s", "",
		"select a given binary from the package, rather than running the default binary")
}

func runPackage(args []string, webmanPkgDir, webmanBinDir, webmanDir string) {
	parts := strings.Split(args[0], "@")
	var pkg string
	var ver string
	var pkgRunFolder string
	var pkgBinDirOrFile string
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

	binPath, err := pkgConf.GetMyBinPath()
	if err != nil {
		exitPrint(1, color.RedString(err.Error()))
	}

	// Is custom version
	var pkgDirName string
	if ver != "" {
		pkgDirName = fmt.Sprintf("%s-%s", pkg, ver)
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
		pkgDirName = *usingVersion
	}
	pkgRunFolder = filepath.Join(webmanPkgDir, pkg, pkgDirName)
	pkgBinDirOrFile = filepath.Join(pkgRunFolder, binPath)
	// default executable to run is the package name
	binName := pkg
	if selectBin != "" {
		binName = selectBin
	}
	// Is folder, pkgBinFile
	pkgBinFileInfo, err := os.Stat(pkgBinDirOrFile)
	if err != nil {
		if os.IsNotExist(err) {
			if runtime.GOOS == "windows" {
				if selectBin != "" {
					exitPrint(1, color.RedString("bin path for package is a file, "+
						"so cannot use --select-bin flag to select a different binary"))
				}
				entries, err := os.ReadDir(filepath.Dir(pkgBinDirOrFile))
				if err != nil {
					IsNotExist(pkg)
				}
				found := false
				for _, entry := range entries {
					if strings.HasPrefix(entry.Name(), binName) {
						found = true
						pkgBinDirOrFile += filepath.Ext(entry.Name())
						break
					}
				}
				if !found {
					IsNotExist(pkg)
				}
			} else {
				IsNotExist(pkg)
			}
		} else {
			exitPrint(1, color.RedString(err.Error()))
		}
	} else if pkgBinFileInfo.IsDir() { // dir
		pkgBinDirOrFile = filepath.Join(pkgBinDirOrFile, binName)
	} else { // non-windows file
		if selectBin != "" {
			exitPrint(1, color.RedString("bin path for package is a file,"+
				"so cannot use --select-bin flag to select a different binary"))
		}
	}

	appCmd := exec.Command(pkgBinDirOrFile, argsApp...)
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

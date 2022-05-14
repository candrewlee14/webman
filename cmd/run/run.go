package run

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"webman/pkgparse"
	"webman/utils"

	"github.com/fatih/color"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var RunCmd = &cobra.Command{
	Use:   "run [pkg](:[binary]) [args...]",
	Short: "run installed packages",
	Long: `
The "run" subcommand runs the installed package binary with the name of the package by default,
or a binary name given after the colon.`,
	Example: `webman run go
webman run bat [FILE]
webman run go@18.0.0
webman run node@17.0.0 --version
webman run node@17.0.0:npm --version
webman run node:npm --version`,
	Run: func(cmd *cobra.Command, args []string) {
		utils.Init()
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
		runPackage(args)

	},
	DisableFlagParsing: true,
}

func init() {
	// RunCmd.Flags().StringVarP(&selectBin, "select-bin", "s", "",
	// 	"select a given binary from the package, rather than running the default binary")
}

func runPackage(args []string) {
	var pkg string
	var ver string
	var binName string
	var initialBinName string
	var pkgRunFolder string
	var pkgBinDirOrFile string
	var argsApp []string

	// Version information
	pkgVerAndBinParts := strings.Split(args[0], ":")
	if len(pkgVerAndBinParts) == 1 {
		pkgStr, verStr, err := utils.ParsePkgVer(args[0])
		if err != nil {
			exitPrint(1, color.RedString(err.Error()))
		}
		pkg = pkgStr
		ver = verStr
	} else if len(pkgVerAndBinParts) == 2 {
		pkgStr, verStr, err := utils.ParsePkgVer(pkgVerAndBinParts[0])
		if err != nil {
			exitPrint(1, color.RedString(err.Error()))
		}
		pkg = pkgStr
		ver = verStr
		binName = pkgVerAndBinParts[1]
		initialBinName = pkgVerAndBinParts[1]
	} else {
		exitPrint(1, "Expected command in form of 'pkg@ver', 'pkg:bin', or 'pkg@ver:bin'")
	}
	// Add args for pkg
	if len(args) > 1 {
		argsApp = args[1:]
	}

	pkgConf, err := pkgparse.ParsePkgConfigLocal(pkg, false)
	if err != nil {
		exitPrint(1, color.RedString(err.Error()))
	}

	binPaths, err := pkgConf.GetMyBinPaths()
	if err != nil {
		exitPrint(1, color.RedString(err.Error()))
	}

	// Is custom version
	var pkgDirName string
	if ver != "" {
		pkgDirName = fmt.Sprintf("%s-%s", pkg, ver)
	} else { // Default version
		usingVersion, err := pkgparse.CheckUsing(pkg)
		if err != nil {
			panic(err)
		}
		if usingVersion == nil {
			exitPrint(1, fmt.Sprintf("Not currently using any %s version\n", color.CyanString(pkg)))
		}
		pkgDirName = *usingVersion
	}
	pkgRunFolder = filepath.Join(utils.WebmanPkgDir, pkg, pkgDirName)
	if _, err = os.Stat(pkgRunFolder); err != nil {
		if os.IsNotExist(err) {
			IsNotExist(pkg, ver)
		}
		exitPrint(1, color.RedString("Error when accessing package version folder: %v\n",
			err))
	}
	var truePkgBinPath *string
	for _, binPath := range binPaths {
		pkgBinDirOrFile = filepath.Join(pkgRunFolder, binPath)
		if binName == "" {
			// default binary name is name of package
			binName = pkg
		}
		// Is folder, pkgBinFile
		pkgBinFileInfo, err := os.Stat(pkgBinDirOrFile)
		if err != nil {
			if os.IsNotExist(err) {
				// at this point, this is either a nonexistent folder
				// or a binary file with an extension we don't yet know
				if runtime.GOOS == "windows" {
					entries, err := os.ReadDir(filepath.Dir(pkgBinDirOrFile))
					if err != nil {
						IsNotExist(pkg, ver)
					}
					for _, entry := range entries {
						eName := entry.Name()
						entryStem := eName[:len(eName)-len(filepath.Ext(eName))]
						if entryStem == binName {
							pkgBinDirOrFile += filepath.Ext(entry.Name())
							pkgBinFileInfo, err = os.Stat(pkgBinDirOrFile)
							if err != nil {
								exitPrint(1, color.RedString("Unable to access binary at %s", pkgBinDirOrFile))
							}
							truePkgBinPath = &pkgBinDirOrFile
							break
						}
					}
				}
			} else {
				exitPrint(1, color.RedString(err.Error()))
			}
		}
		if pkgBinFileInfo.IsDir() { // dir
			pkgBinDirOrFile = filepath.Join(pkgBinDirOrFile, binName)
			if _, err = os.Stat(pkgBinDirOrFile); err != nil {
				if !os.IsNotExist(err) {
					exitPrint(1, color.RedString("Error when accessing binary: %v\n",
						err))
				}
			} else {
				truePkgBinPath = &pkgBinDirOrFile
			}
		} else if initialBinName != "" { // is a file
			exitPrint(1, color.RedString("bin path for package is a file,"+
				"so cannot select a different binary"))
		} else {
			truePkgBinPath = &pkgBinDirOrFile
		}
		if truePkgBinPath != nil {
			break
		}
	}

	if truePkgBinPath == nil {
		exitPrint(1, "No "+color.CyanString(binName)+" binary exists for "+
			color.CyanString(pkgDirName))
	}
	appCmd := exec.Command(*truePkgBinPath, argsApp...)
	appCmd.Stderr = os.Stderr
	appCmd.Stdout = os.Stdout
	appCmd.Stdin = os.Stdin
	appCmd.Env = os.Environ()

	// Start package
	if err := appCmd.Run(); err != nil {
		if os.IsNotExist(err) {
			IsNotExist(pkg, ver)
		}
		exitPrint(1, color.RedString(err.Error()))
	}

}

func IsNotExist(pkg string, ver string) {
	fmt.Printf("No versions of %s@%s are currently installed.\n", color.CyanString(pkg), color.MagentaString(ver))
	os.Exit(1)
}

func exitPrint(code int, text string) {
	fmt.Println(text)
	os.Exit(code)
}

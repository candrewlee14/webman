package run

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/candrewlee14/webman/config"
	"github.com/candrewlee14/webman/pkgparse"
	"github.com/candrewlee14/webman/utils"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return runPackage(args)
	},
	DisableFlagParsing: true,
}

func runPackage(args []string) error {
	var pkg string
	var ver string
	var binName string
	var initialBinName string
	var pkgRunFolder string
	var pkgBinDirOrFile string
	var argsApp []string

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Version information
	pkgVerAndBinParts := strings.Split(args[0], ":")
	if len(pkgVerAndBinParts) == 1 {
		pkgStr, verStr, err := utils.ParsePkgVer(args[0])
		if err != nil {
			return err
		}
		pkg = pkgStr
		ver = verStr
	} else if len(pkgVerAndBinParts) == 2 {
		pkgStr, verStr, err := utils.ParsePkgVer(pkgVerAndBinParts[0])
		if err != nil {
			return err
		}
		pkg = pkgStr
		ver = verStr
		binName = pkgVerAndBinParts[1]
		initialBinName = pkgVerAndBinParts[1]
	} else {
		return fmt.Errorf("Expected command in form of 'pkg@ver', 'pkg:bin', or 'pkg@ver:bin'")
	}
	// Add args for pkg
	if len(args) > 1 {
		argsApp = args[1:]
	}

	pkgConf, err := pkgparse.ParsePkgConfigLocal(cfg.PkgRepos, pkg)
	if err != nil {
		return err
	}

	binPaths, err := pkgConf.GetMyBinPaths()
	if err != nil {
		return err
	}

	// Is custom version
	var pkgDirName string
	if ver != "" {
		pkgDirName = fmt.Sprintf("%s-%s", pkg, ver)
	} else { // Default version
		usingVersion, err := pkgparse.CheckUsing(pkg)
		if err != nil {
			return err
		}
		if usingVersion == nil {
			return fmt.Errorf("Not currently using any %s version\n", pkg)
		}
		pkgDirName = *usingVersion
	}
	pkgRunFolder = filepath.Join(utils.WebmanPkgDir, pkg, pkgDirName)
	if _, err = os.Stat(pkgRunFolder); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("No versions of %s@%s are currently installed.\n", pkg, ver)
		}
		return fmt.Errorf("Error when accessing package version folder: %v\n", err)
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
				if utils.GOOS == "windows" {
					entries, err := os.ReadDir(filepath.Dir(pkgBinDirOrFile))
					if err != nil {
						return fmt.Errorf("No versions of %s@%s are currently installed.\n", pkg, ver)
					}
					for _, entry := range entries {
						eName := entry.Name()
						entryStem := eName[:len(eName)-len(filepath.Ext(eName))]
						if entryStem == binName {
							pkgBinDirOrFile += filepath.Ext(entry.Name())
							pkgBinFileInfo, err = os.Stat(pkgBinDirOrFile)
							if err != nil {
								return fmt.Errorf("Unable to access binary at %s", pkgBinDirOrFile)
							}
							truePkgBinPath = &pkgBinDirOrFile
							break
						}
					}
				}
			} else {
				return err
			}
		}
		if pkgBinFileInfo.IsDir() { // dir
			var binExt string
			if utils.GOOS == "windows" {
				binExt = ".exe"
			}
			pkgBinDirOrFile = filepath.Join(pkgBinDirOrFile, binName+binExt)
			if _, err = os.Stat(pkgBinDirOrFile); err != nil {
				if !os.IsNotExist(err) {
					return fmt.Errorf("Error when accessing binary: %v\n", err)
				}
			} else {
				truePkgBinPath = &pkgBinDirOrFile
			}
		} else if initialBinName != "" { // is a file
			return fmt.Errorf("bin path for package is a file, so cannot select a different binary")
		} else {
			truePkgBinPath = &pkgBinDirOrFile
		}
		if truePkgBinPath != nil {
			break
		}
	}

	if truePkgBinPath == nil {
		return fmt.Errorf("No " + binName + " binary exists for " +
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
			return fmt.Errorf("No versions of %s@%s are currently installed.\n", pkg, ver)
		}
		return err
	}
	return nil
}

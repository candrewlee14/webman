package bintest

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"webman/cmd/add"
	"webman/cmd/dev/check"
	"webman/multiline"
	"webman/pkgparse"
	"webman/utils"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var OsOptions []string = []string{"windows", "darwin", "linux"}
var ArchOptions []string = []string{"amd64", "arm64"}

// CheckCmd represents the remove command
var BintestCmd = &cobra.Command{
	Use:   "bintest [pkg]",
	Short: "Test the installation & binary paths for each platform for a package",
	Long: `
The "bintest" tests that binary paths given in a package recipe have valid binaries, and displays them.`,
	Example: `webman bintest zoxide -l ~/repos/webman-pkgs/`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Help()
			os.Exit(0)
		}
		utils.Init()
		homedir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		pkg := args[0]
		var pairResults map[string]bool = map[string]bool{}
		if err := check.CheckPkgConfig(pkg); err != nil {
			color.Red("Pkg Config Error: %v", err)
		}
		pkgConf, err := pkgparse.ParsePkgConfigLocal(pkg, true)
	osLoop:
		for _, osStr := range OsOptions {
			//Example: convert "windows" GOOS to "win" pkgOS
			osPkgStr := pkgparse.GOOStoPkgOs[osStr]
		archLoop:
			for _, arch := range ArchOptions {
				osPairStr := fmt.Sprintf("%s-%s", osStr, arch)
				if err != nil {
					color.Red("Error: %v", err)
					os.Exit(1)
				}
				if _, osSupported := pkgConf.OsMap[osPkgStr]; !osSupported {
					color.HiBlack("Skipping all %s: unsupported by pkg", osStr)
					continue osLoop
				}
				if _, archSupported := pkgConf.ArchMap[arch]; !archSupported {
					color.HiBlack("Skipping %s-%s: unsupported by pkg", osStr, arch)
					continue archLoop
				}
				for _, pair := range pkgConf.Ignore {
					if pair.Arch == arch && pair.Os == osPkgStr {
						color.HiBlack("Skipping %s-%s: pair ignored by pkg", osStr, arch)
						continue archLoop
					}
				}
				InitTestDir(osStr, arch, homedir)
				color.HiCyan("Trying OS=%s Arch=%s installation", osStr, arch)
				fmt.Println("Putting installation in ", utils.WebmanDir)
				var wg sync.WaitGroup
				ml := multiline.New(len(args), os.Stdout)
				wg.Add(1)
				pairResults[osPairStr] = add.InstallPkg(pkg, 0, 1, &wg, &ml)
			}
		}
		allSucceed := true
		fmt.Println("\nResults:")
		for key, val := range pairResults {
			if val {
				color.Green("  %s : SUCCESS", key)
			} else {
				allSucceed = false
				color.Red("  %s : FAIL", key)
			}
		}
		if allSucceed {
			color.HiGreen("\nAll supported OSs & Arches for %s have valid installs!", pkg)
			os.RemoveAll(filepath.Join(homedir, ".webman", "test"))
		} else {
			color.HiRed("\nSome supported OSs & Arches for %s have invalid installs.", pkg)
		}
	},
}

func InitTestDir(osStr string, arch string, homedir string) {
	utils.WebmanDir = filepath.Join(homedir, ".webman", "test", osStr, arch)
	utils.WebmanPkgDir = filepath.Join(utils.WebmanDir, "/pkg")
	utils.WebmanBinDir = filepath.Join(utils.WebmanDir, "/bin")
	utils.WebmanTmpDir = filepath.Join(utils.WebmanDir, "/tmp")
	// leave WebmanRecipesDir the way it was

	if err := os.MkdirAll(utils.WebmanBinDir, os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(utils.WebmanPkgDir, os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(utils.WebmanTmpDir, os.ModePerm); err != nil {
		panic(err)
	}
	utils.GOOS = osStr
	utils.GOARCH = arch
}

func init() {

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// removeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

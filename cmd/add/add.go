package add

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
	"webman/multiline"
	"webman/pkgparse"

	"github.com/fatih/color"

	progressbar "github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var doRefresh bool
var recipeDir string

// AddCmd represents the add command
var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "install packages",
	Long: `
The "add" subcommand installs packages.`,
	Example: `webman add go
webman add go@18.0.0
webman add go zig rg
webman add go@18.0.0 zig@9.1.0 rg@13.0.0`,
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
		webmanTmpDir := filepath.Join(webmanDir, "/tmp")
		defer os.RemoveAll(webmanTmpDir)
		if err = os.MkdirAll(webmanBinDir, os.ModePerm); err != nil {
			panic(err)
		}
		if err = os.MkdirAll(webmanPkgDir, os.ModePerm); err != nil {
			panic(err)
		}
		if err = os.MkdirAll(webmanTmpDir, os.ModePerm); err != nil {
			panic(err)
		}
		shouldRefresh, err := pkgparse.ShouldRefreshRecipes(webmanDir)
		if err != nil {
			panic(err)
		}
		if shouldRefresh || doRefresh {
			color.HiBlue("Refreshing package recipes")
			if err = pkgparse.RefreshRecipes(webmanDir); err != nil {
				fmt.Println(err)
			}
		}
		var wg sync.WaitGroup
		ml := multiline.New(len(args), os.Stdout)
		wg.Add(len(args))
		success := true
		for i, arg := range args {
			i := i
			arg := arg
			go func() {
				if !installPkg(arg, i, len(args), webmanDir, &wg, &ml) {
					success = false
				}
			}()
		}
		wg.Wait()
		if !success {
			color.Magenta("Not all packages installed successfully")
			os.Exit(1)
		}
		color.Green("All packages installed successfully!")
	},
}

func init() {
	AddCmd.Flags().BoolVar(&doRefresh, "refresh", false, "force refresh of package recipes")
	AddCmd.Flags().StringVarP(&recipeDir, "local-recipes", "l", "", "use given local recipe directory")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func cleanUpFailedInstall(webmanDir string, pkg string, extractPath string) {
	os.RemoveAll(extractPath) // clean up failed installation
	pkgDir := filepath.Join(webmanDir, "pkg", pkg)
	dirs, err := os.ReadDir(pkgDir)
	if err == nil && len(dirs) == 0 {
		os.RemoveAll(pkgDir)
	}
}

func DownloadUrl(url string, f io.Writer, pkg string, ver string, argNum int, argCount int, ml *multiline.MultiLogger) bool {
	r, err := http.Get(url)
	ml.Printf(argNum, "Downloading file at %s", url)
	if err != nil {
		ml.Printf(argNum, color.RedString("%v", err))
		return false
	}
	defer r.Body.Close()
	if !(r.StatusCode >= 200 && r.StatusCode < 300) {
		switch r.StatusCode {
		case 404, 403:
			ml.Printf(argNum, color.RedString("unable to find %s@%s on the web at %s", pkg, ver, url))
		default:
			ml.Printf(argNum, color.RedString("bad HTTP Response: %s", r.Status))
		}
		return false
	}
	bar := progressbar.NewOptions64(r.ContentLength,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetWriter(ioutil.Discard),
		progressbar.OptionShowBytes(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetDescription(
			fmt.Sprintf("[cyan][%d/%d][reset] Downloading [cyan]"+pkg+"[reset] file...", argNum+1, argCount)),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
	go func() {
		for !bar.IsFinished() {
			barStr := bar.String()
			ml.Printf(argNum, "%s", barStr)
			time.Sleep(100 * time.Millisecond)
		}
	}()
	if _, err = io.Copy(io.MultiWriter(f, bar), r.Body); err != nil {
		ml.Printf(argNum, color.RedString("%v", err))
		return false
	}
	return true
}

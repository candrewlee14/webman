package add

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/candrewlee14/webman/link"
	"github.com/candrewlee14/webman/multiline"
	"github.com/candrewlee14/webman/pkgparse"
	"github.com/candrewlee14/webman/utils"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"golang.org/x/sync/errgroup"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var doRefresh bool
var switchFlag bool

// addCmd represents the add command
var AddCmd = &cobra.Command{
	Use:   "add [pkgs...]",
	Short: "install packages",
	Long: `
The "add" subcommand installs packages.`,
	Example: `webman add go
webman add go@18.0.0
webman add go zig rg
webman add go@18.0.0 zig@9.1.0 rg@13.0.0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.Init()
		if len(args) == 0 {
			cmd.Help()
			return nil
		}
		defer os.RemoveAll(utils.WebmanTmpDir)
		// if local recipe flag is not set
		if utils.RecipeDirFlag == "" {
			// only refresh if not using local
			shouldRefresh, err := pkgparse.ShouldRefreshRecipes()
			if err != nil {
				return err
			}
			if shouldRefresh || doRefresh {
				color.HiBlue("Refreshing package recipes")
				if err = pkgparse.RefreshRecipes(); err != nil {
					color.Red("%v", err)
				}
			}
		}
		if !InstallAllPkgs(args) {
			return fmt.Errorf("Not all packages installed successfully")
		}
		color.Green("All %d packages are installed!", len(args))
		return nil
	},
}

func init() {
	AddCmd.Flags().BoolVar(&doRefresh, "refresh", false, "force refresh of package recipes")
	AddCmd.Flags().BoolVar(&switchFlag, "switch", false, "switch to use this new package version")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func cleanUpFailedInstall(pkg string, extractPath string) {
	os.RemoveAll(extractPath)
	pkgDir := filepath.Join(utils.WebmanPkgDir, pkg)
	dirs, err := os.ReadDir(pkgDir)
	if err == nil && len(dirs) == 0 {
		os.RemoveAll(pkgDir)
	}
}

func DownloadUrl(url string, filePath string, pkg string, ver string, argNum int, argCount int, ml *multiline.MultiLogger) bool {
	f, err := os.OpenFile(filePath,
		os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		ml.Printf(argNum, color.RedString("%v", err))
		return false
	}
	defer f.Close()

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
	colorOn := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	saucer := "[green]▅[reset]"
	saucerHead := "[green]▅[reset]"
	saucerPadding := "[light_gray]▅[reset]"
	barStart := ""
	barEnd := ""
	barDesc := fmt.Sprintf("[cyan][%d/%d][reset] Downloading [cyan]"+pkg+"[reset] file...", argNum+1, argCount)
	if !colorOn {
		saucer = "="
		saucerHead = ">"
		saucerPadding = " "
		barDesc = fmt.Sprintf("[%d/%d] Downloading "+pkg+" file...", argNum+1, argCount)
		barStart = "["
		barEnd = "]"
	}
	ansiOn := isatty.IsTerminal(os.Stdout.Fd())
	bar := progressbar.NewOptions64(r.ContentLength,
		progressbar.OptionEnableColorCodes(colorOn),
		progressbar.OptionUseANSICodes(ansiOn),
		progressbar.OptionSetWriter(ioutil.Discard),
		progressbar.OptionShowBytes(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetDescription(barDesc),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        saucer,
			SaucerHead:    saucerHead,
			SaucerPadding: saucerPadding,
			BarStart:      barStart,
			BarEnd:        barEnd,
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

func CreateLinks(pkg string, stem string, confBinPaths []string) (bool, error) {
	binPaths, linkPaths, err := link.GetBinPathsAndLinkPaths(pkg, stem, confBinPaths)
	if err != nil {
		return false, err
	}

	var eg errgroup.Group
	for i, linkPath := range linkPaths {
		binPath := binPaths[i]
		linkPath := linkPath // this supresses the warning for linkPath closure capture
		eg.Go(func() error {
			didLink, err := link.AddLink(binPath, linkPath)
			if err != nil {
				return err
			}
			if !didLink {
				return fmt.Errorf("failed to create link to %s", binPath)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return false, err
	}
	if err = pkgparse.WriteUsing(pkg, stem); err != nil {
		panic(err)
	}
	return true, nil
}

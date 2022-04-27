package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"webman/pkgparse"
	"webman/unpack"
	progressbar "github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
    "sync"
)

func installPkg(arg string, argNum int, argCount int, webmanDir string, wg *sync.WaitGroup) {
    defer wg.Done()
    webmanPkgDir := filepath.Join(webmanDir, "/pkg")
    //webmanBinDir := filepath.Join(webmanDir, "/bin")
    webmanTmpDir := filepath.Join(webmanDir, "/tmp")
    parts := strings.Split(arg, "@")
    var pkg string
    var ver string
    var pkgConf pkgparse.PkgConfig
    if len(parts) == 1 {
        pkg = parts[0]
    } else if len(parts) == 2 {
        pkg = parts[0]
        ver = parts[1]
    } else {
        panic("Packages should be in format 'pkg' or 'pkg@version'")
    }
    pkgConf = pkgparse.ParsePkgConfig(pkg)
    if len(ver) == 0 {
        fmt.Println("Finding latest version tag")
        ver = pkgConf.GetLatestVersion()
        fmt.Println("Found version tag: ", ver)
    }
    stem, ext, url := pkgConf.GetAssetStemExtUrl(ver)
    fileName := stem + "." + ext
    downloadPath := filepath.Join(webmanTmpDir, fileName)
    
    extractPath := filepath.Join(webmanPkgDir, pkg, stem)
    // If file exists
    if _, err := os.Stat(extractPath); !os.IsNotExist(err) {
        fmt.Println(pkg, "@", ver, "is already installed!")
        os.Exit(0)
    }
    fmt.Println(downloadPath)
    f, err := os.OpenFile(downloadPath, 
        os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        panic(err)
    }
    defer f.Close()

    r, err := http.Get(url)
    if err != nil {
        panic(err)
    }
    defer r.Body.Close()

    bar := progressbar.NewOptions64(r.ContentLength,
        progressbar.OptionEnableColorCodes(true),
        progressbar.OptionShowBytes(true),
        progressbar.OptionFullWidth(),
        progressbar.OptionSetDescription(
            fmt.Sprintf("[cyan][%d/%d][reset] Downloading [cyan]" + pkg + "[reset] file...", argNum, argCount)),
        progressbar.OptionThrottle(20 * time.Millisecond),
        progressbar.OptionSetTheme(progressbar.Theme{
            Saucer:        "[green]=[reset]",
            SaucerHead:    "[green]>[reset]",
            SaucerPadding: " ",
            BarStart:      "[",
            BarEnd:        "]",
        }))
    _, err = io.Copy(io.MultiWriter(f, bar), r.Body)
    fmt.Println("")
    if err != nil {
        panic(err)
    }
    err = unpack.Unpack(downloadPath, filepath.Join(webmanPkgDir, pkg), ext)
    if err != nil {
        panic(err)
    }
    
}

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
        homeDir, err := os.UserHomeDir();
        if err != nil {
           panic(err);
        }
        webmanDir := filepath.Join(homeDir, "/.webman"); 
        webmanPkgDir := filepath.Join(webmanDir, "/pkg")
        webmanBinDir := filepath.Join(webmanDir, "/bin")
        webmanTmpDir := filepath.Join(webmanDir, "/tmp")
        defer os.RemoveAll(webmanTmpDir)
        err = os.MkdirAll(webmanBinDir, os.ModePerm)
        if err != nil {
           panic(err);
        }
        err = os.MkdirAll(webmanPkgDir, os.ModePerm)
        if err != nil {
           panic(err);
        }
        err = os.MkdirAll(webmanTmpDir, os.ModePerm)
        if err != nil {
           panic(err);
        }
        var wg sync.WaitGroup
        for i, arg := range args {
            wg.Add(1)
            go installPkg(arg, i, len(args), webmanDir, &wg)
        }
        wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

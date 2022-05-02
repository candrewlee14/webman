package add

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"webman/link"
	"webman/multiline"
	"webman/pkgparse"
	"webman/unpack"

	"github.com/fatih/color"
)

func installPkg(arg string, argNum int, argCount int, webmanDir string, wg *sync.WaitGroup, ml *multiline.MultiLogger) bool {
	defer wg.Done()
	webmanPkgDir := filepath.Join(webmanDir, "/pkg")
	webmanTmpDir := filepath.Join(webmanDir, "/tmp")
	parts := strings.Split(arg, "@")
	var pkg string
	var ver string
	if len(parts) == 1 {
		pkg = parts[0]
	} else if len(parts) == 2 {
		pkg = parts[0]
		ver = parts[1]
	} else {
		ml.Printf(argNum, "Packages should be in format 'pkg' or 'pkg@version'")
		return false
	}
	if len(ver) == 0 {
		ml.SetPrefix(argNum, color.YellowString(pkg)+": ")

	} else {
		ml.SetPrefix(argNum, color.YellowString(pkg)+"@"+color.YellowString(ver)+": ")
	}
	foundRecipe := make(chan bool)
	ml.PrintUntilDone(argNum,
		fmt.Sprintf("Finding package recipe for %s", color.CyanString(pkg)),
		foundRecipe,
		500,
	)
	if recipeDir == "" {
		recipeDir = filepath.Join(webmanDir, "recipes")
	}
	recipeDir, err := filepath.Abs(recipeDir)
	if err != nil {
		ml.Printf(argNum, "Failed converting local package directory to absolute path")
		return false
	}
	pkgConf, err := pkgparse.ParsePkgConfigLocal(recipeDir, pkg)
	foundRecipe <- true
	if err != nil {
		ml.Printf(argNum, color.RedString("%v", err))
		return false
	}
	for _, ignorePair := range pkgConf.Ignore {
		if runtime.GOOS == ignorePair.Os && runtime.GOARCH == ignorePair.Arch {
			ml.Printf(argNum, color.RedString("unsupported OS + Arch for this package"))
			return false
		}
	}
	if len(ver) == 0 {
		foundLatest := make(chan bool)
		ml.PrintUntilDone(argNum,
			fmt.Sprintf("Finding latest %s version tag", color.CyanString(pkg)),
			foundLatest,
			500,
		)
		verPtr, err := pkgConf.GetLatestVersion()
		foundLatest <- true
		if err != nil {
			ml.Printf(argNum, color.RedString("unable to find latest version tag: %v", err))
			return false
		}
		ver = *verPtr
		ml.Printf(argNum, "Found %s version tag: %s", color.CyanString(pkg), color.MagentaString(ver))
	}
	stemPtr, extPtr, urlPtr, err := pkgConf.GetAssetStemExtUrl(ver)
	if err != nil {
		ml.Printf(argNum, color.RedString("%v", err))
		return false
	}
	stem := *stemPtr
	ext := *extPtr
	url := *urlPtr
	fileName := stem + "." + ext
	downloadPath := filepath.Join(webmanTmpDir, fileName)

	extractStem := fmt.Sprintf("%s-%s", pkg, ver)
	extractPath := filepath.Join(webmanPkgDir, pkg, extractStem)
	// If file exists
	if _, err := os.Stat(extractPath); !os.IsNotExist(err) {
		ml.Printf(argNum, color.HiBlackString("Already installed!"))
		return true
	}
	f, err := os.OpenFile(downloadPath,
		os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		ml.Printf(argNum, color.RedString("%v", err))
		return false
	}
	defer f.Close()
	if !DownloadUrl(url, f, pkg, ver, argNum, argCount, ml) {
		return false
	}
	hasUnpacked := make(chan bool)
	ml.PrintUntilDone(argNum,
		fmt.Sprintf("Unpacking %s.%s", stem, ext),
		hasUnpacked,
		500,
	)
	err = unpack.Unpack(downloadPath, webmanDir, pkg, extractStem, ext, pkgConf.ExtractHasRoot)
	hasUnpacked <- true
	if err != nil {
		ml.Printf(argNum, color.RedString("%v", err))
		cleanUpFailedInstall(webmanDir, pkg, extractPath)
		return false
	}
	ml.Printf(argNum, "Completed unpacking %s@%s", color.CyanString(pkg), color.MagentaString(ver))

	using, err := pkgparse.CheckUsing(pkg, webmanDir)
	if err != nil {
		cleanUpFailedInstall(webmanDir, pkg, extractPath)
		panic(err)
	}
	if using == nil {
		binPath, err := pkgConf.GetMyBinPath()
		if err != nil {
			cleanUpFailedInstall(webmanDir, pkg, extractPath)
			ml.Printf(argNum, color.RedString("%v", err))
			return false
		}
		madeLinks, err := link.CreateLinks(webmanDir, pkg, extractStem, binPath)
		if err != nil {
			cleanUpFailedInstall(webmanDir, pkg, extractPath)
			ml.Printf(argNum, color.RedString("Failed creating links: %v", err))
			return false
		}
		if !madeLinks {
			cleanUpFailedInstall(webmanDir, pkg, extractPath)
			ml.Printf(argNum, color.RedString("Failed creating links"))
			return false
		}
		ml.Printf(argNum, "Now using %s@%s", color.CyanString(pkg), color.MagentaString(ver))
	}
	ml.Printf(argNum, color.GreenString("Successfully installed!"))
	return true
}

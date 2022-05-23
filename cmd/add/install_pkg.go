package add

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/candrewlee14/webman/link"
	"github.com/candrewlee14/webman/multiline"
	"github.com/candrewlee14/webman/pkgparse"
	"github.com/candrewlee14/webman/unpack"
	"github.com/candrewlee14/webman/utils"

	"github.com/fatih/color"
)

func InstallAllPkgs(args []string) bool {
	var wg sync.WaitGroup
	ml := multiline.New(len(args), os.Stdout)
	wg.Add(len(args))
	results := make(chan bool, len(args))
	for i, arg := range args {
		i := i
		arg := arg
		go func() {
			res := InstallPkg(arg, i, len(args), &wg, &ml)
			results <- res
		}()
	}
	wg.Wait()
	success := true
	for i := 0; i < len(args); i++ {
		success = success && <-results
	}
	return success
}

func InstallPkg(arg string, argIndex int, argCount int, wg *sync.WaitGroup, ml *multiline.MultiLogger) bool {
	defer wg.Done()
	pkg, ver, err := utils.ParsePkgVer(arg)
	if err != nil {
		ml.Printf(argIndex, color.RedString(err.Error()))
		return false
	}
	if len(ver) == 0 {
		ml.SetPrefix(argIndex, color.CyanString(pkg)+": ")

	} else {
		ml.SetPrefix(argIndex, color.CyanString(pkg)+"@"+color.CyanString(ver)+": ")
	}
	foundRecipe := make(chan bool)
	ml.PrintUntilDone(argIndex,
		fmt.Sprintf("Finding package recipe for %s", color.CyanString(pkg)),
		foundRecipe,
		500,
	)
	pkgConf, err := pkgparse.ParsePkgConfigLocal(pkg, false)
	foundRecipe <- true
	if err != nil {
		ml.Printf(argIndex, color.RedString("%v", err))
		return false
	}
	pkgOS := pkgparse.GOOStoPkgOs[utils.GOOS]
	for _, ignorePair := range pkgConf.Ignore {
		if pkgOS == ignorePair.Os && utils.GOARCH == ignorePair.Arch {
			ml.Printf(argIndex, color.RedString("unsupported OS + Arch for this package"))
			return false
		}
	}
	if len(ver) == 0 || pkgConf.ForceLatest {
		foundLatest := make(chan bool)
		ml.PrintUntilDone(argIndex,
			fmt.Sprintf("Finding latest %s version tag", color.CyanString(pkg)),
			foundLatest,
			500,
		)
		verPtr, err := pkgConf.GetLatestVersion()
		foundLatest <- true
		if err != nil {
			ml.Printf(argIndex, color.RedString("unable to find latest version tag: %v", err))
			return false
		}
		if pkgConf.ForceLatest && len(ver) != 0 && *verPtr != ver {
			ml.Printf(argIndex, color.RedString("This package requires using the latest version, which is currently %s",
				color.MagentaString(*verPtr)))
			return false
		}
		ver = *verPtr
		ml.Printf(argIndex, "Found %s version tag: %s", color.CyanString(pkg), color.MagentaString(ver))
	}
	stemPtr, extPtr, urlPtr, err := pkgConf.GetAssetStemExtUrl(ver)
	if err != nil {
		ml.Printf(argIndex, color.RedString("%v", err))
		return false
	}
	stem := *stemPtr
	ext := *extPtr
	url := *urlPtr

	fileName := stem
	if ext != "" {
		fileName += "." + ext
	}
	downloadPath := filepath.Join(utils.WebmanTmpDir, fileName)

	extractStem := utils.CreateStem(pkg, ver)
	extractPath := filepath.Join(utils.WebmanPkgDir, pkg, extractStem)

	// If file exists
	if _, err := os.Stat(extractPath); !os.IsNotExist(err) {
		ml.Printf(argIndex, color.HiBlackString("Already installed!"))
		return true
	}
	if pkgConf.IsBinary && utils.GOOS == "windows" {
		url += ".exe"
	}
	if !DownloadUrl(url, downloadPath, pkg, ver, argIndex, argCount, ml) {
		return false
	}
	if pkgConf.IsBinary {
		if err = os.Chmod(downloadPath, 0755); err != nil {
			ml.Printf(argIndex, color.RedString("Failed to make download executable!"))
			return false
		}
		if err = os.MkdirAll(extractPath, os.ModePerm); err != nil {
			ml.Printf(argIndex, color.RedString("Failed to create package-version path!"))
			return false
		}
		binPath := filepath.Join(extractPath, pkgConf.Title)
		if utils.GOOS == "windows" {
			binPath += ".exe"
		}
		if err = os.Rename(downloadPath, binPath); err != nil {
			ml.Printf(argIndex, color.RedString("Failed to rename temporary download to new path!"))
			return false
		}
	} else {
		hasUnpacked := make(chan bool)
		ml.PrintUntilDone(argIndex,
			fmt.Sprintf("Unpacking %s.%s", stem, ext),
			hasUnpacked,
			500,
		)
		var extractHasRoot bool
		if m, ok := pkgConf.OsMap[pkgOS]; ok {
			extractHasRoot = m.ExtractHasRoot
		}
		err = unpack.Unpack(downloadPath, pkg, extractStem, extractHasRoot)
		hasUnpacked <- true
		if err != nil {
			ml.Printf(argIndex, color.RedString("%v", err))
			cleanUpFailedInstall(pkg, extractPath)
			return false
		}
		ml.Printf(argIndex, "Completed unpacking %s@%s", color.CyanString(pkg), color.MagentaString(ver))
	}
	using, err := pkgparse.CheckUsing(pkg)
	if err != nil {
		cleanUpFailedInstall(pkg, extractPath)
		panic(err)
	}
	if using == nil || switchFlag {
		binPaths, err := pkgConf.GetMyBinPaths()
		if err != nil {
			cleanUpFailedInstall(pkg, extractPath)
			ml.Printf(argIndex, color.RedString("%v", err))
			return false
		}
		madeLinks, err := link.CreateLinks(pkg, ver, binPaths)
		if err != nil {
			cleanUpFailedInstall(pkg, extractPath)
			ml.Printf(argIndex, color.RedString("Failed creating links: %v", err))
			return false
		}
		if !madeLinks {
			cleanUpFailedInstall(pkg, extractPath)
			ml.Printf(argIndex, color.RedString("Failed creating links"))
			return false
		}
		ml.Printf(argIndex, "Now using %s@%s", color.CyanString(pkg), color.MagentaString(ver))
	}
	ml.Printf(argIndex, color.GreenString("Successfully installed!"))
	return true
}

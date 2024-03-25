package add

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/candrewlee14/webman/cmd/remove"
	"github.com/candrewlee14/webman/config"
	"github.com/candrewlee14/webman/link"
	"github.com/candrewlee14/webman/multiline"
	"github.com/candrewlee14/webman/pkgparse"
	"github.com/candrewlee14/webman/unpack"
	"github.com/candrewlee14/webman/utils"

	"github.com/fatih/color"
)

type PkgInstallResult struct {
	Name    string
	Ver     string
	PkgConf *pkgparse.PkgConfig
}

func InstallAllPkgs(pkgRepos []*config.PkgRepo, args []string, removeOld bool, switchFlag bool) []PkgInstallResult {
	var wg sync.WaitGroup
	ml := multiline.New(len(args), os.Stdout)
	wg.Add(len(args))
	results := make(chan *PkgInstallResult, len(args))
	for i, arg := range args {
		i := i
		arg := arg
		go func() {
			res := InstallPkg(pkgRepos, arg, i, len(args), &wg, &ml, removeOld, switchFlag)
			results <- res
		}()
	}
	wg.Wait()
	pkgs := make([]PkgInstallResult, 0, len(args))
	for i := 0; i < len(args); i++ {
		res := <-results
		if res != nil {
			pkgs = append(pkgs, *res)
		}
	}
	return pkgs
}

func InstallPkg(
	pkgRepos []*config.PkgRepo,
	arg string, argIndex int, argCount int,
	wg *sync.WaitGroup, ml *multiline.MultiLogger,
	removeOld bool,
	switchFlag bool,
) *PkgInstallResult {
	defer wg.Done()
	pkg, ver, err := utils.ParsePkgVer(arg)
	if err != nil {
		ml.Printf(argIndex, color.RedString(err.Error()))
		return nil
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
		50,
	)
	pkgConf, err := pkgparse.ParsePkgConfigLocal(pkgRepos, pkg)
	foundRecipe <- true
	if err != nil {
		ml.Printf(argIndex, color.RedString("%v", err))
		return nil
	}
	pkgOS := pkgparse.GOOStoPkgOs[utils.GOOS]
	for _, ignorePair := range pkgConf.Ignore {
		if pkgOS == ignorePair.Os && utils.GOARCH == ignorePair.Arch {
			ml.Printf(argIndex, color.RedString("unsupported OS + Arch for this package"))
			return nil
		}
	}
	if len(ver) == 0 || pkgConf.ForceLatest {
		foundLatest := make(chan bool)
		ml.PrintUntilDone(argIndex,
			fmt.Sprintf("Finding latest %s version tag", color.CyanString(pkg)),
			foundLatest,
			50,
		)
		verPtr, err := pkgConf.GetLatestVersion()
		foundLatest <- true
		if err != nil {
			ml.Printf(argIndex, color.RedString("unable to find latest version tag: %v", err))
			return nil
		}
		if pkgConf.ForceLatest && len(ver) != 0 && *verPtr != ver {
			ml.Printf(argIndex, color.RedString("This package requires using the latest version, which is currently %s",
				color.MagentaString(*verPtr)))
			return nil
		}
		ver = *verPtr
		ml.Printf(argIndex, "Found %s version tag: %s", color.CyanString(pkg), color.MagentaString(ver))
	}
	stemPtr, extPtr, urlPtr, err := pkgConf.GetAssetStemExtUrl(ver)
	if err != nil {
		ml.Printf(argIndex, color.RedString("%v", err))
		return nil
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
		return &PkgInstallResult{pkg, ver, pkgConf}
	}
	if !DownloadUrl(url, downloadPath, pkg, ver, argIndex, argCount, ml) {
		return nil
	}
	var isRawBinary bool
	if m, ok := pkgConf.OsMap[pkgOS]; ok {
		isRawBinary = m.IsRawBinary
	}
	if isRawBinary {
		if err = os.Chmod(downloadPath, 0o755); err != nil {
			ml.Printf(argIndex, color.RedString("Failed to make download executable!"))
			return nil
		}
		if err = os.MkdirAll(extractPath, os.ModePerm); err != nil {
			ml.Printf(argIndex, color.RedString("Failed to create package-version path!"))
			return nil
		}
		binPath := filepath.Join(extractPath, pkgConf.Title)
		if utils.GOOS == "windows" {
			binPath += ".exe"
		}
		if err = os.Rename(downloadPath, binPath); err != nil {
			ml.Printf(argIndex, color.RedString("Failed to rename temporary download to new path!"))
			return nil
		}
	} else {
		hasUnpacked := make(chan bool)
		ml.PrintUntilDone(argIndex,
			fmt.Sprintf("Unpacking %s.%s", stem, ext),
			hasUnpacked,
			50,
		)
		var extractHasRoot bool
		if m, ok := pkgConf.OsMap[pkgOS]; ok {
			extractHasRoot = m.ExtractHasRoot
		}
		err = unpack.Unpack(downloadPath, pkg, extractStem, extractHasRoot)
		hasUnpacked <- true
		if err != nil {
			ml.Printf(argIndex, color.RedString("%v", err))
			CleanUpFailedInstall(pkg, extractPath)
			return nil
		}
		ml.Printf(argIndex, "Completed unpacking %s@%s", color.CyanString(pkg), color.MagentaString(ver))
	}
	using, err := pkgparse.CheckUsing(pkg)
	if err != nil {
		CleanUpFailedInstall(pkg, extractPath)
		panic(err)
	}
	if using != nil {
		if removeOld {
			if err = remove.RemovePkgVer(*using, using, pkg, pkgConf); err != nil {
				ml.Printf(argIndex, color.RedString("Failed to remove old version: %v", err))
			} else {
				ml.Printf(argIndex, "Removed old version %s", color.CyanString(*using))
			}
		}
	}
	if using == nil || switchFlag {
		binPaths, err := pkgConf.GetMyBinPaths()
		if err != nil {
			CleanUpFailedInstall(pkg, extractPath)
			ml.Printf(argIndex, color.RedString("%v", err))
			return nil
		}
		renames, err := pkgConf.GetRenames()
		if err != nil {
			ml.Printf(argIndex, color.RedString("Failed creating links: %v", err))
			return nil
		}
		madeLinks, err := link.CreateLinks(pkg, ver, binPaths, renames)
		if err != nil {
			CleanUpFailedInstall(pkg, extractPath)
			ml.Printf(argIndex, color.RedString("Failed creating links: %v", err))
			return nil
		}
		if !madeLinks {
			CleanUpFailedInstall(pkg, extractPath)
			ml.Printf(argIndex, color.RedString("Failed creating links"))
			return nil
		}
		ml.Printf(argIndex, "Now using %s@%s", color.CyanString(pkg), color.MagentaString(ver))
	}

	ml.Printf(argIndex, color.GreenString("Successfully installed!"))
	if p, err := exec.LookPath(pkg); err == nil && !strings.Contains(p, utils.WebmanBinDir) {
		ml.Printf(argIndex, color.YellowString("Found another binary at %q that may interfere", p))
	}
	return &PkgInstallResult{pkg, ver, pkgConf}
}

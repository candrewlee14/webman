package link

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/candrewlee14/webman/pkgparse"
	"github.com/candrewlee14/webman/utils"

	"golang.org/x/sync/errgroup"
)

func GetBinPathsAndLinkPaths(
	pkg string,
	ver string,
	confBinPaths []string,
	renames []pkgparse.RenameItem,
) ([]string, []string, error) {
	var binPaths []string
	var linkPaths []string
	binExt := ""
	if utils.GOOS == "windows" {
		binExt = ".exe"
	}
	stem := utils.CreateStem(pkg, ver)
	for _, confBinPath := range confBinPaths {
		binPath := filepath.Join(utils.WebmanPkgDir, pkg, stem, confBinPath+binExt)
		fileInfo, err := os.Stat(binPath)
		// If config binary path points to a file
		if err == nil && !fileInfo.IsDir() {
			linkPath := GetLinkPathIfExec(binPath, renames)
			if linkPath != nil {
				binPaths = append(binPaths, binPath)
				linkPaths = append(linkPaths, *linkPath)
			}
		} else {
			binDir := filepath.Join(utils.WebmanPkgDir, pkg, stem, confBinPath)
			binDirEntries, err := os.ReadDir(binDir)
			if err != nil {
				return []string{}, []string{}, err
			}
			for _, entry := range binDirEntries {
				if !entry.Type().IsDir() {
					binPath := filepath.Join(binDir, entry.Name())
					linkPath := GetLinkPathIfExec(binPath, renames)
					if linkPath != nil {
						binPaths = append(binPaths, binPath)
						linkPaths = append(linkPaths, *linkPath)
					}
				}
			}
		}
		if len(linkPaths) == 0 {
			return []string{}, []string{}, fmt.Errorf("given binary path had no executable files")
		}
	}

	return binPaths, linkPaths, nil
}

// Returns a link path to ~/.webman/bin/foo
// This is system-agnostic, it will always be that format
func GetLinkPathIfExec(binPath string, renames []pkgparse.RenameItem) *string {
	binFile := filepath.Base(binPath)
	binName := binFile[:len(binFile)-len(filepath.Ext(binFile))]
	for _, r := range renames {
		binName = strings.ReplaceAll(binName, r.From, r.To)
	}
	linkPath := filepath.Join(utils.WebmanBinDir, binName)
	fi, err := os.Stat(binPath)
	if err != nil {
		return nil
	}
	// If not executable
	if !(fi.Mode()&0o111 != 0) && runtime.GOOS != "windows" {
		return nil
	}
	return &linkPath
}

// Create a link to an old file at the new path
func AddLink(old string, new string) (bool, error) {
	if err := os.Remove(new); err != nil {
		// if the file did exist and it's a different error, return it
		if !os.IsNotExist(err) {
			return false, err
		}
	}
	if err := os.Symlink(old, new); err != nil {
		return false, err
	}
	return true, nil
}

func CreateLinks(pkg string, ver string, confBinPaths []string, renames []pkgparse.RenameItem) (bool, error) {
	binPaths, linkPaths, err := GetBinPathsAndLinkPaths(pkg, ver, confBinPaths, renames)
	if err != nil {
		return false, err
	}

	var eg errgroup.Group
	for i, linkPath := range linkPaths {
		binPath := binPaths[i]
		linkPath := linkPath // this suppresses the warning for linkPath closure capture
		eg.Go(func() error {
			didLink, err := AddLink(binPath, linkPath)
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
	if err = pkgparse.WriteUsing(pkg, utils.CreateStem(pkg, ver)); err != nil {
		panic(err)
	}
	return true, nil
}

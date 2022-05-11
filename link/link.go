package link

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"webman/pkgparse"
	"webman/utils"

	"golang.org/x/sync/errgroup"
)

func GetBinPathsAndLinkPaths(
	pkg string,
	stem string,
	confBinPaths []string,
) ([]string, []string, error) {
	var binPaths []string
	var linkPaths []string
	binExt := ""
	if runtime.GOOS == "windows" {
		binExt = ".exe"
	}
	for _, confBinPath := range confBinPaths {

		binPath := filepath.Join(utils.WebmanPkgDir, pkg, stem, confBinPath+binExt)
		fileInfo, err := os.Stat(binPath)
		// If config binary path points to a file
		if err == nil && !fileInfo.IsDir() {
			linkPath := GetLinkPathIfExec(binPath)
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
					linkPath := GetLinkPathIfExec(binPath)
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
func GetLinkPathIfExec(binPath string) *string {
	binFile := filepath.Base(binPath)
	binName := binFile[:len(binFile)-len(filepath.Ext(binFile))]
	linkPath := filepath.Join(utils.WebmanBinDir, binName)
	if runtime.GOOS == "windows" {
		switch filepath.Ext(binPath) {
		case ".bat", ".exe", ".cmd":
			break
		// if not an executable filetype
		default:
			return nil
		}
	} else {
		fi, err := os.Stat(binPath)
		if err != nil {
			return nil
		}
		// If not executable
		if !(fi.Mode()&0111 != 0) {
			return nil
		}
	}
	return &linkPath
}

// Create a link to an old file at the new path
// On windows, .bat will be appended to the new path to make a batch file
func AddLink(old string, new string) (bool, error) {
	if runtime.GOOS == "windows" {
		f, err := os.Create(new + ".bat")
		if err != nil {
			return false, err
		}
		defer f.Close()
		_, err = f.WriteString(
			fmt.Sprintf("@echo off\n%s", old) + ` %*`,
		)
		if err != nil {
			return false, err
		}
	} else {
		if err := os.Remove(new); err != nil {
			// if the file did exist and it's a different error, return it
			if !os.IsNotExist(err) {
				return false, err
			}
		}
		if err := os.Symlink(old, new); err != nil {
			return false, err
		}
	}
	return true, nil
}

func CreateLinks(pkg string, stem string, confBinPaths []string) (bool, error) {
	binPaths, linkPaths, err := GetBinPathsAndLinkPaths(pkg, stem, confBinPaths)
	if err != nil {
		return false, err
	}

	var eg errgroup.Group
	for i, linkPath := range linkPaths {
		binPath := binPaths[i]
		linkPath := linkPath // this supresses the warning for linkPath closure capture
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
	if err = pkgparse.WriteUsing(pkg, stem); err != nil {
		panic(err)
	}
	return true, nil
}

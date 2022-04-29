package link

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func GetBinPathsAndLinkPaths(argNum int,
	webmanDir string,
	pkg string,
	stem string,
	confBinPath string,
) ([]string, []string, error) {
	var binPaths []string
	var linkPaths []string
	binExt := ""
	if runtime.GOOS == "windows" {
		binExt = ".exe"
	}
	binPath := filepath.Join(webmanDir, "pkg", pkg, stem, confBinPath+binExt)
	f, err := os.Open(binPath)
	// If config binary path points to a file
	if err == nil {
		linkPath := GetLinkPathIfExec(binPath, webmanDir)
		if linkPath != nil {
			binPaths = append(binPaths, binPath)
			linkPaths = append(linkPaths, *linkPath)
		}
	} else {
		f.Close()
		binDir := filepath.Join(webmanDir, "pkg", pkg, stem, confBinPath)
		binDirEntries, err := os.ReadDir(binDir)
		if err != nil {
			return []string{}, []string{}, err
		}
		for _, entry := range binDirEntries {
			if !entry.Type().IsDir() {
				binPath := filepath.Join(binDir, entry.Name())
				linkPath := GetLinkPathIfExec(binPath, webmanDir)
				if linkPath != nil {
					binPaths = append(binPaths, binPath)
					linkPaths = append(linkPaths, *linkPath)
				}
			}
		}
	}
	return binPaths, linkPaths, nil
}

// Returns a link path to ~/.webman/bin/foo
// This is system-agnostic, it will always be that format
func GetLinkPathIfExec(binPath string, webmanDir string) *string {
	binFile := filepath.Base(binPath)
	binName := binFile[:len(binFile)-len(filepath.Ext(binFile))]
	linkPath := filepath.Join(webmanDir, "bin", binName)
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
		err := os.Symlink(old, new)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

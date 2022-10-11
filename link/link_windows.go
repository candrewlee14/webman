//go:build windows
// +build windows

package link

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/candrewlee14/webman/utils"
)

// If symlink creation fails, fallbacks to creating a script named `${executable name minus extension}.cmd`.
func symlink(old string, new string) (bool, error) {
	err := os.Symlink(old, new)
	if err == nil {
		return true, nil
	}

	f, err := os.Create(cmdPath(new))
	if err != nil {
		return false, err
	}
	defer f.Close()
	_, err = f.WriteString(
		fmt.Sprintf("@echo off\r\n%s %s\r\n", old, `%*`),
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func cmdPath(orig string) string {
	noext := orig[:len(orig)-len(filepath.Ext(orig))]
	return utils.LinkName(noext)
}

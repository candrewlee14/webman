//go:build darwin || freebsd || linux || netbsd || openbsd
// +build darwin freebsd linux netbsd openbsd

package link

import "os"

func symlink(old string, new string) (bool, error) {
	err := os.Symlink(old, new)
	if err != nil {
		return false, err
	}

	return true, nil
}

//go:build windows
// +build windows

package utils

import "fmt"

func LinkName(pkg string) string {
	return fmt.Sprintf("%s.cmd", pkg)
}

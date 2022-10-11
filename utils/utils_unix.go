//go:build darwin || freebsd || linux || netbsd || openbsd
// +build darwin freebsd linux netbsd openbsd

package utils

func LinkName(pkg string) string {
	return pkg
}

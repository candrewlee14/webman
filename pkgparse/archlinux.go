package pkgparse

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
)

type ArchLinuxPkgInfo struct {
	PkgVer string
}

func getLatestArchLinuxPkgVersion(archpkg string) (*ArchLinuxPkgInfo, error) {
	url := "https://raw.githubusercontent.com/archlinux/svntogit-community/master/" +
		archpkg + "/trunk/PKGBUILD"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Host", "raw.githubusercontent.com")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	if !(r.StatusCode >= 200 && r.StatusCode < 300) {
		return nil, fmt.Errorf("Bad HTTP Response: " + r.Status)
	}
	scanner := bufio.NewScanner(r.Body)
	var pkgInfo ArchLinuxPkgInfo
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "pkgver=") {
			pkgInfo.PkgVer = scanner.Text()[7:]
			return &pkgInfo, nil
		}
	}

	return nil, fmt.Errorf("invalid PKGBUILD file with no pkgver field")
}

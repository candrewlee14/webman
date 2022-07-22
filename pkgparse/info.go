package pkgparse

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/candrewlee14/webman/utils"

	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

type PkgInfo struct {
	Title   string
	Tagline string
	About   string
}

func ParsePkgInfo(pkgRepo, pkg string) (*PkgInfo, error) {
	pkgConfPath := filepath.Join(pkgRepo, "pkgs", pkg+utils.PkgRecipeExt)
	dat, err := os.ReadFile(pkgConfPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no package recipe exists for %s", pkg)
		}
		return nil, err
	}
	var pkgInfo PkgInfo
	if err = yaml.Unmarshal(dat, &pkgInfo); err != nil {
		return nil, fmt.Errorf("unable to parse package recipe for %s: %v", pkg, err)
	}
	pkgInfo.Title = pkg
	return &pkgInfo, nil
}

func ParseMultiPkgInfo(pkgs []string) ([]PkgInfo, error) {
	var m sync.Mutex
	pkgInfos := make([]PkgInfo, len(pkgs))
	var eg errgroup.Group
	for i, pkg := range pkgs {
		i := i
		pkg := pkg
		eg.Go(func() error {
			pkgInfo, err := ParsePkgInfo("", pkg)
			if err != nil {
				return nil
			}
			m.Lock()
			pkgInfos[i] = *pkgInfo
			m.Unlock()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return []PkgInfo{}, err
	}
	return pkgInfos, nil
}

package pkgparse

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"webman/utils"

	"github.com/go-yaml/yaml"
	"golang.org/x/sync/errgroup"
)

type PkgInfo struct {
	Title   string
	Tagline string
	About   string
}

func ParsePkgInfo(pkg string) (*PkgInfo, error) {
	pkgConfPath := filepath.Join(utils.WebmanRecipeDir, "pkgs", pkg+".yaml")
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
			pkgInfo, err := ParsePkgInfo(pkg)
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

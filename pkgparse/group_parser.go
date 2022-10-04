package pkgparse

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/candrewlee14/webman/config"
	"github.com/candrewlee14/webman/utils"

	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

type PkgGroupConfig struct {
	Title   string
	Tagline string
	About   string

	InfoUrl  string   `yaml:"info_url"`
	Packages []string `yaml:"packages"`
}

func ParseGroupConfig(r io.Reader, name string) (*PkgGroupConfig, error) {
	var groupConf PkgGroupConfig
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read package group file: %v", err)
	}
	if err = yaml.Unmarshal(data, &groupConf); err != nil {
		return nil, fmt.Errorf("invalid format for package group: %v", err)
	}
	if len(groupConf.Packages) == 0 {
		return nil, fmt.Errorf("no packages in package group %s", color.YellowString(name))
	}
	return &groupConf, nil
}

func ParseGroupConfigLocal(pkgRepos []*config.PkgRepo, group string) (*PkgGroupConfig, string, error) {
	var groupConfPath string
	var repo string
	for _, pkgRepo := range pkgRepos {
		groupPath := filepath.Join(pkgRepo.Path(), "groups", group+utils.GroupRecipeExt)
		_, err := os.Stat(groupPath)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return nil, "", err
			}
			continue
		}
		groupConfPath = groupPath
		repo = pkgRepo.Path()
		break
	}
	if groupConfPath == "" {
		return nil, "", fmt.Errorf("no package group exists for %s", group)
	}

	fi, err := os.Open(groupConfPath)
	if err != nil {
		return nil, "", err
	}
	defer fi.Close()

	groupCfg, err := ParseGroupConfig(fi, group)
	return groupCfg, repo, err
}

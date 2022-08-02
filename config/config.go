package config

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/candrewlee14/webman/schema"
	"github.com/candrewlee14/webman/utils"

	"github.com/mholt/archiver/v3"
	"gopkg.in/yaml.v3"
)

//go:embed config.yaml
var defaultConfig []byte

type Config struct {
	RefreshInterval time.Duration `yaml:"refresh_interval"`
	PkgRepos        []*PkgRepo    `yaml:"pkg_repos"`
}

type PkgRepoType string

const (
	PkgRepoTypeGitHub PkgRepoType = "github"
	PkgRepoTypeGitea  PkgRepoType = "gitea"
)

type PkgRepo struct {
	Name   string      `yaml:"name"`
	Type   PkgRepoType `yaml:"type"`
	User   string      `yaml:"user"`
	Repo   string      `yaml:"repo"`
	Branch string      `yaml:"branch"`

	GiteaURL string `yaml:"gitea_url"`
}

func (p PkgRepo) Path() string {
	return filepath.Join(utils.WebmanRecipeDir, p.Name)
}

func (p PkgRepo) ShouldRefreshRecipes(refreshInterval time.Duration) (bool, error) {
	fi, err := os.Stat(p.Path())
	if err != nil {
		return false, err
	}
	return time.Since(fi.ModTime()) > refreshInterval, nil
}

func (p PkgRepo) RefreshRecipes() error {
	var url string
	switch p.Type {
	case PkgRepoTypeGitHub:
		url = fmt.Sprintf("https://github.com/%s/%s/archive/refs/heads/%s.tar.gz", p.User, p.Repo, p.Branch)
	case PkgRepoTypeGitea:
		url = fmt.Sprintf("%s/api/v1/repos/%s/%s/archive/%s.tar.gz", p.GiteaURL, p.User, p.Repo, p.Branch)
	default:
		return errors.New("unknown package repository type")
	}

	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if !(r.StatusCode >= 200 && r.StatusCode < 300) {
		return fmt.Errorf("Bad HTTP Response: %s", r.Status)
	}

	if err = os.RemoveAll(p.Path()); err != nil {
		return err
	}
	if err = os.MkdirAll(utils.WebmanTmpDir, os.ModePerm); err != nil {
		return err
	}
	tmpZipFile, err := os.CreateTemp(utils.WebmanTmpDir, "recipes-*.tar.gz")
	if err != nil {
		return err
	}
	if _, err = io.Copy(tmpZipFile, r.Body); err != nil {
		return err
	}

	tmpRecipeDir := filepath.Join(utils.WebmanTmpDir, "recipes")
	if err = archiver.Unarchive(tmpZipFile.Name(), tmpRecipeDir); err != nil {
		return err
	}
	fdir, err := os.ReadDir(tmpRecipeDir)
	if err != nil {
		return err
	}
	if len(fdir) != 1 {
		return fmt.Errorf("expected unzipped refresh to have a single root folder")
	}
	innerTmpFolder := filepath.Join(tmpRecipeDir, fdir[0].Name())
	if err = os.Rename(innerTmpFolder, p.Path()); err != nil {
		return err
	}

	return nil
}

func Load() (*Config, error) {
	if utils.RecipeDirFlag != "" {
		// local only
		utils.WebmanRecipeDir = utils.RecipeDirFlag
		return &Config{
			RefreshInterval: 0,
			PkgRepos: []*PkgRepo{
				{Name: "."},
			},
		}, nil
	}

	fi, err := os.Open(utils.WebmanConfig)
	if err != nil {
		// If it doesn't exist, write out the default
		if errors.Is(err, os.ErrNotExist) {
			if err := writeDefaultConfig(); err != nil {
				return nil, err
			}
			return Load()
		}
		return nil, err
	}
	defer fi.Close()

	var buf bytes.Buffer
	tee := io.TeeReader(fi, &buf)
	if err := schema.LintConfig(tee); err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.NewDecoder(&buf).Decode(&cfg); err != nil {
		return nil, err
	}
	for _, pkgRepo := range cfg.PkgRepos {
		if pkgRepo.Branch == "" {
			pkgRepo.Branch = "main"
		}
	}

	return &cfg, nil
}

func writeDefaultConfig() error {
	fi, err := os.Create(utils.WebmanConfig)
	if err != nil {
		return err
	}
	defer fi.Close()

	_, err = fi.Write(defaultConfig)
	return err
}

package config

import (
	_ "embed"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/candrewlee14/webman/utils"

	"gopkg.in/yaml.v3"
)

//go:embed default.yml
var defaultConfig []byte

type Config struct {
	RefreshInterval time.Duration `yaml:"refresh_interval"`
	PkgRepos        []PkgRepo     `yaml:"pkg_repos"`
}

type RefreshFile struct {
	LastUpdated time.Time
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
	refreshFileDir := filepath.Join(p.Path(), utils.RefreshFileName)
	data, err := os.ReadFile(refreshFileDir)
	if err != nil {
		// if err occurred and file does exist
		if !os.IsNotExist(err) {
			return false, err
		}
	}
	var refreshFile RefreshFile
	if err = yaml.Unmarshal(data, &refreshFile); err != nil {
		return true, err
	}
	return time.Since(refreshFile.LastUpdated) > refreshInterval, nil
}

func (p PkgRepo) RefreshRecipes() error {
	switch p.Type {
	case PkgRepoTypeGitHub:
		return p.githubRefresh()
	case PkgRepoTypeGitea:
		return p.giteaRefresh()
	default:
		return errors.New("unknown package repository type")
	}
}

func Load() (*Config, error) {
	if utils.RecipeDirFlag != "" {
		// local only
		utils.WebmanRecipeDir = utils.RecipeDirFlag
		return &Config{
			RefreshInterval: 0,
			PkgRepos: []PkgRepo{
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

	var cfg Config
	if err := yaml.NewDecoder(fi).Decode(&cfg); err != nil {
		return nil, err
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

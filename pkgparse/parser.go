package pkgparse

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-yaml/yaml"
)

type OsInfo struct {
	Name string `yaml:"name"`
	Ext  string `yaml:"ext"`
}

type PkgConfig struct {
	InfoUrl         string `yaml:"info_url"`
	ReleasesUrl     string `yaml:"releases_url"`
	BaseDownloadUrl string `yaml:"base_download_url"`
	GitUser         string `yaml:"git_user"`
	GitRepo         string `yaml:"git_repo"`
	SourceUrl       string `yaml:"source_url"`

	FilenameFormat   string `yaml:"filename_format"`
	VersionType      string `yaml:"version_type"`
	LatestStrategy   string `yaml:"latest_strategy"`
	ArchLinuxPkgName string `yaml:"arch_linux_pkg_name"`

	BinPath        string `yaml:"bin_path"`
	ExtractHasRoot bool   `yaml:"extract_has_root"`

	OsMap   map[string]OsInfo `yaml:"os_map"`
	ArchMap map[string]string `yaml:"arch_map"`
}

func ParsePkgConfig(pkg string) (*PkgConfig, error) {
	curDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("unable to get working dir: %v", err)
	}
	pkgConfPath := filepath.Join(curDir, "/pkgs/"+pkg+".yaml")
	dat, err := os.ReadFile(pkgConfPath)
	if err != nil {
		return nil, fmt.Errorf("unable find package recipe for %s", pkg)
	}
	var pkgConf PkgConfig
	if err = yaml.UnmarshalStrict(dat, &pkgConf); err != nil {
		return nil, fmt.Errorf("unable parse package recipe for %s: %v", pkg, err)
	}
	return &pkgConf, nil
}

var goOsToPkgOs = map[string]string{
	"darwin":  "macos",
	"windows": "win",
	"linux":   "linux",
}

func (pkgConf *PkgConfig) GetLatestVersion() (*string, error) {
	switch pkgConf.LatestStrategy {
	case "github-release":
		rel, err := getLatestGithubReleaseTag(pkgConf.GitUser, pkgConf.GitRepo)
		if err != nil {
			return nil, err
		}
		return &rel.TagName, nil
	case "arch-linux-community":
		rel, err := getLatestArchLinuxPkgVersion(pkgConf.ArchLinuxPkgName)
		if err != nil {
			return nil, err
		}
		return &rel.PkgVer, nil
	}
	return nil, fmt.Errorf("no implemented latest version resolution strategy for %q",
		pkgConf.LatestStrategy)
}

///
func (pkgConf *PkgConfig) GetAssetStemExtUrl(version string) (*string, *string, *string, error) {
	pkgOs, exists := goOsToPkgOs[runtime.GOOS]
	if !exists {
		return nil, nil, nil, fmt.Errorf("unsupported operating system")
	}
	osInf, exists := pkgConf.OsMap[pkgOs]
	if !exists {
		return nil, nil, nil, fmt.Errorf("package has no binary for operating system: %s", pkgOs)
	}
	archStr, exists := pkgConf.ArchMap[runtime.GOARCH]
	if !exists {
		return nil, nil, nil, fmt.Errorf("package has no binary for architecture: %s", archStr)
	}
	baseUrl := pkgConf.BaseDownloadUrl
	baseUrl = strings.ReplaceAll(baseUrl, "[VER]", version)
	baseUrl = strings.ReplaceAll(baseUrl, "[OS]", osInf.Name)
	baseUrl = strings.ReplaceAll(baseUrl, "[ARCH]", archStr)
	baseUrl = strings.ReplaceAll(baseUrl, "[EXT]", osInf.Ext)

	fileStem := pkgConf.FilenameFormat
	fileStem = strings.ReplaceAll(fileStem, "[VER]", version)
	fileStem = strings.ReplaceAll(fileStem, "[OS]", osInf.Name)
	fileStem = strings.ReplaceAll(fileStem, "[ARCH]", archStr)
	fileStem = strings.ReplaceAll(fileStem, ".[EXT]", "")
	stem := baseUrl + fileStem + "." + osInf.Ext
	return &fileStem, &osInf.Ext, &stem, nil
}

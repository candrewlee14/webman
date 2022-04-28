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

	BinPathFormat  string `yaml:"bin_path_format"`
	ExtractHasRoot bool   `yaml:"extract_has_root"`

	OsMap   map[string]OsInfo `yaml:"os_map"`
	ArchMap map[string]string `yaml:"arch_map"`
}

func ParsePkgConfig(pkg string) PkgConfig {
	curDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	pkgConfPath := filepath.Join(curDir, "/pkgs/"+pkg+".yaml")
	dat, err := os.ReadFile(pkgConfPath)
	if err != nil {
		panic(err)
	}
	var pkgConf PkgConfig
	err = yaml.UnmarshalStrict(dat, &pkgConf)
	if err != nil {
		panic(err)
	}
	return pkgConf
}

var goOsToPkgOs = map[string]string{
	"darwin":  "macos",
	"windows": "win",
	"linux":   "linux",
}

func (pkgConf *PkgConfig) GetLatestVersion() string {
	switch pkgConf.LatestStrategy {
	case "github-release":
		return getLatestGithubReleaseTag(pkgConf.GitUser, pkgConf.GitRepo).TagName
	case "arch-linux-community":
		return getLatestArchLinuxPkgVersion(pkgConf.ArchLinuxPkgName).PkgVer
	}
	panic(fmt.Sprintf("No implemented latest version resolution strategy for %q",
		pkgConf.LatestStrategy))
}

///
func (pkgConf *PkgConfig) GetAssetStemExtUrl(version string) (string, string, string) {
	pkgOs, exists := goOsToPkgOs[runtime.GOOS]
	if !exists {
		panic("Unsupported operating system")
	}
	osInf, exists := pkgConf.OsMap[pkgOs]
	if !exists {
		panic(fmt.Sprintf("Package has no binary for operating system: %s", pkgOs))
	}
	archStr, exists := pkgConf.ArchMap[runtime.GOARCH]
	if !exists {
		panic(fmt.Sprintf("Package has no binary for architecture: %s", archStr))
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
	return fileStem, osInf.Ext, baseUrl + fileStem + "." + osInf.Ext
}

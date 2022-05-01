package pkgparse

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-yaml/yaml"
)

type UsingInfo struct {
	Using string
}

type OsInfo struct {
	Name    string `yaml:"name"`
	Ext     string `yaml:"ext"`
	BinPath string `yaml:"bin_path"`
}

type PkgConfig struct {
	InfoUrl         string `yaml:"info_url"`
	ReleasesUrl     string `yaml:"releases_url"`
	BaseDownloadUrl string `yaml:"base_download_url"`
	GitUser         string `yaml:"git_user"`
	GitRepo         string `yaml:"git_repo"`
	SourceUrl       string `yaml:"source_url"`

	FilenameFormat   string `yaml:"filename_format"`
	LatestStrategy   string `yaml:"latest_strategy"`
	ArchLinuxPkgName string `yaml:"arch_linux_pkg_name"`

	ExtractHasRoot bool `yaml:"extract_has_root"`

	OsMap   map[string]OsInfo `yaml:"os_map"`
	ArchMap map[string]string `yaml:"arch_map"`
}

var goOsToPkgOs = map[string]string{
	"darwin":  "macos",
	"windows": "win",
	"linux":   "linux",
}

func (pkgConf *PkgConfig) GetMyBinPath() (string, error) {
	osStr, exists := goOsToPkgOs[runtime.GOOS]
	if !exists {
		return "", fmt.Errorf("unsupported OS")
	}
	osInfo, exists := pkgConf.OsMap[osStr]
	if !exists {
		return "", fmt.Errorf("package does not support this OS")
	}
	return osInfo.BinPath, nil
}

// Check using file.
// If using.yaml file doesn't exist, it is not using anything
func CheckUsing(pkg string, webmanDir string) (*string, error) {
	usingPath := filepath.Join(webmanDir, "pkg", pkg, "using.yaml")
	usingContent, err := os.ReadFile(usingPath)
	if err != nil {
		return nil, nil
	}
	var usingInfo UsingInfo
	if err = yaml.UnmarshalStrict(usingContent, &usingInfo); err != nil {
		return nil, err
	}
	return &usingInfo.Using, nil
}

func WriteUsing(pkg string, webmanDir string, using string) error {
	usingInfo := UsingInfo{
		Using: using,
	}
	data, err := yaml.Marshal(usingInfo)
	if err != nil {
		return err
	}
	usingPath := filepath.Join(webmanDir, "pkg", pkg, "using.yaml")
	if err := os.WriteFile(usingPath, data, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func RemoveUsing(pkg string, webmanDir string) error {
	usingPath := filepath.Join(webmanDir, "pkg", pkg, "using.yaml")
	if err := os.Remove(usingPath); err != nil {
		return err
	}
	return nil
}

func ParsePkgConfigOnline(pkg string) (*PkgConfig, error) {
	pkgConfUrl := "https://raw.githubusercontent.com/candrewlee14/webman-pkgs/main/pkgs/" + pkg + ".yaml"
	r, err := http.Get(pkgConfUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to download %s package recipe: %v", pkg, err)
	}
	defer r.Body.Close()
	if !(r.StatusCode >= 200 && r.StatusCode < 300) {
		switch r.StatusCode {
		case 404:
			return nil, fmt.Errorf("no package recipe for %q exists", pkg)
		default:
			return nil, fmt.Errorf(
				"bad HTTP response when downloading package recipe for %q: %s", pkg, r.Status)
		}
	}
	dat, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to download %s package recipe: %v", pkg, err)
	}
	var pkgConf PkgConfig
	if err = yaml.UnmarshalStrict(dat, &pkgConf); err != nil {
		return nil, fmt.Errorf("unable parse package recipe for %s: %v", pkg, err)
	}
	return &pkgConf, nil
}

func ParsePkgConfigLocal(recipeDir string, pkg string) (*PkgConfig, error) {
	pkgConfPath := filepath.Join(recipeDir, "pkgs", pkg+".yaml")
	dat, err := os.ReadFile(pkgConfPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no package recipe exists for %s", pkg)
		}
		return nil, err
	}
	var pkgConf PkgConfig
	if err = yaml.UnmarshalStrict(dat, &pkgConf); err != nil {
		return nil, fmt.Errorf("unable parse package recipe for %s: %v", pkg, err)
	}
	return &pkgConf, nil
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
		return nil, nil, nil, fmt.Errorf("package has no binary for architecture: %s", runtime.GOARCH)
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

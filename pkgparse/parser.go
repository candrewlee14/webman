package pkgparse

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"webman/utils"

	"github.com/go-yaml/yaml"
)

type UsingInfo struct {
	Using string
}

type OsInfo struct {
	Name     string        `yaml:"name"`
	Ext      string        `yaml:"ext"`
	BinPaths SingleOrMulti `yaml:"bin_path"`
}

type OsArchPair struct {
	Os   string `yaml:"os"`
	Arch string `yaml:"arch"`
}

type PkgConfig struct {
	Title   string
	Tagline string
	About   string

	InfoUrl         string `yaml:"info_url"`
	ReleasesUrl     string `yaml:"releases_url"`
	BaseDownloadUrl string `yaml:"base_download_url"`
	GitUser         string `yaml:"git_user"`
	GitRepo         string `yaml:"git_repo"`
	SourceUrl       string `yaml:"source_url"`

	FilenameFormat   string `yaml:"filename_format"`
	VersionFormat    string `yaml:"version_format"`
	LatestStrategy   string `yaml:"latest_strategy"`
	ForceLatest      bool   `yaml:"force_latest"`
	AllowPrerelease  bool   `yaml:"allow_prerelease"`
	ArchLinuxPkgName string `yaml:"arch_linux_pkg_name"`

	IsBinary       bool `yaml:"is_binary"`
	ExtractHasRoot bool `yaml:"extract_has_root"`

	OsMap   map[string]OsInfo `yaml:"os_map"`
	ArchMap map[string]string `yaml:"arch_map"`
	Ignore  []OsArchPair      `yaml:"ignore"`
}

var GOOStoPkgOs = map[string]string{
	"darwin":  "macos",
	"windows": "win",
	"linux":   "linux",
}

type SingleOrMulti struct {
	Values []string
}

func (sm *SingleOrMulti) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var multi []string
	err := unmarshal(&multi)
	if err != nil {
		var single string
		err := unmarshal(&single)
		if err != nil {
			return err
		}
		sm.Values = make([]string, 1)
		sm.Values[0] = single
	} else {
		sm.Values = multi
	}
	return nil
}

func (pkgConf *PkgConfig) GetMyBinPaths() ([]string, error) {
	osStr, exists := GOOStoPkgOs[runtime.GOOS]
	if !exists {
		return []string{}, fmt.Errorf("unsupported OS")
	}
	osInfo, exists := pkgConf.OsMap[osStr]
	if !exists {
		return []string{}, fmt.Errorf("package does not support this OS")
	}
	if pkgConf.IsBinary {
		return []string{pkgConf.Title}, nil
	}
	if len(osInfo.BinPaths.Values) == 0 {
		osInfo.BinPaths = SingleOrMulti{[]string{""}}
	}
	return osInfo.BinPaths.Values, nil
}

// Check using file.
// If using.yaml file doesn't exist, it is not using anything
func CheckUsing(pkg string) (*string, error) {
	usingPath := filepath.Join(utils.WebmanPkgDir, pkg, "using.yaml")
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

func WriteUsing(pkg string, using string) error {
	usingInfo := UsingInfo{
		Using: using,
	}
	data, err := yaml.Marshal(usingInfo)
	if err != nil {
		return err
	}
	usingPath := filepath.Join(utils.WebmanPkgDir, pkg, "using.yaml")
	if err := os.WriteFile(usingPath, data, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func RemoveUsing(pkg string) error {
	usingPath := filepath.Join(utils.WebmanPkgDir, pkg, "using.yaml")
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
		case 404, 403:
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

func ParsePkgConfigLocal(pkg string, strict bool) (*PkgConfig, error) {
	pkgConfPath := filepath.Join(utils.WebmanRecipeDir, "pkgs", pkg+".yaml")
	dat, err := os.ReadFile(pkgConfPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no package recipe exists for %s", pkg)
		}
		return nil, err
	}
	var pkgConf PkgConfig
	if strict {
		if err = yaml.UnmarshalStrict(dat, &pkgConf); err != nil {
			return nil, fmt.Errorf("unable to strict parse package recipe for %s: %v", pkg, err)
		}
	} else {
		if err = yaml.Unmarshal(dat, &pkgConf); err != nil {
			return nil, fmt.Errorf("unable to parse package recipe for %s: %v", pkg, err)
		}
	}
	pkgConf.BaseDownloadUrl = strings.ReplaceAll(pkgConf.BaseDownloadUrl, "[GIT_USER]", pkgConf.GitUser)
	pkgConf.BaseDownloadUrl = strings.ReplaceAll(pkgConf.BaseDownloadUrl, "[GIT_REPO]", pkgConf.GitRepo)

	pkgConf.InfoUrl = strings.ReplaceAll(pkgConf.InfoUrl, "[GIT_USER]", pkgConf.GitUser)
	pkgConf.InfoUrl = strings.ReplaceAll(pkgConf.InfoUrl, "[GIT_REPO]", pkgConf.GitRepo)

	pkgConf.ReleasesUrl = strings.ReplaceAll(pkgConf.InfoUrl, "[GIT_USER]", pkgConf.GitUser)
	pkgConf.ReleasesUrl = strings.ReplaceAll(pkgConf.InfoUrl, "[GIT_REPO]", pkgConf.GitRepo)

	pkgConf.SourceUrl = strings.ReplaceAll(pkgConf.SourceUrl, "[GIT_USER]", pkgConf.GitUser)
	pkgConf.SourceUrl = strings.ReplaceAll(pkgConf.SourceUrl, "[GIT_REPO]", pkgConf.GitRepo)

	return &pkgConf, nil
}

func (pkgConf *PkgConfig) GetLatestVersion() (*string, error) {
	var version string
	switch pkgConf.LatestStrategy {
	case "github-release":
		rel, err := getLatestGithubReleaseTag(pkgConf.GitUser, pkgConf.GitRepo, pkgConf.AllowPrerelease)
		if err != nil {
			return nil, err
		}
		version = rel.TagName
	case "arch-linux-community":
		rel, err := getLatestArchLinuxPkgVersion(pkgConf.ArchLinuxPkgName)
		if err != nil {
			return nil, err
		}
		version = rel.PkgVer
	}
	if version == "" {
		return nil, fmt.Errorf("no implemented latest version resolution strategy for %q",
			pkgConf.LatestStrategy)
	}
	parsedVer, err := ParseVersion(version, pkgConf.VersionFormat)
	if err != nil {
		return nil, fmt.Errorf("unable to parse version: %v", err)
	}
	return parsedVer, nil
}

func ParseVersion(versionStr string, versionFmt string) (*string, error) {
	if versionFmt == "" {
		versionFmt = "[VER]"
	}
	versionMatchExp := strings.Replace(versionFmt, "[VER]", "(.+)", 1)
	exp, err := regexp.Compile(versionMatchExp)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex based on version_format: %v", err)
	}
	matchedVer := exp.FindStringSubmatch(versionStr)
	if len(matchedVer) != 2 || matchedVer[1] == "" {
		return nil, fmt.Errorf("failed to match version %q with given version_format", versionStr)
	}
	return &matchedVer[1], nil
}

///
func (pkgConf *PkgConfig) GetAssetStemExtUrl(version string) (*string, *string, *string, error) {
	pkgOs, exists := GOOStoPkgOs[runtime.GOOS]
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
	dot := ""
	if osInf.Ext != "" {
		dot = "."
	}
	stem := baseUrl + fileStem + dot + osInf.Ext
	return &fileStem, &osInf.Ext, &stem, nil
}

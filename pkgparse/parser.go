package pkgparse

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/candrewlee14/webman/config"
	"github.com/candrewlee14/webman/utils"

	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

type UsingInfo struct {
	Using string
}

type RenameItem struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

type OsInfo struct {
	Name                   string        `yaml:"name"`
	Ext                    string        `yaml:"ext"`
	BinPaths               SingleOrMulti `yaml:"bin_path"`
	ExtractHasRoot         bool          `yaml:"extract_has_root"`
	IsRawBinary            bool          `yaml:"is_raw_binary"`
	FilenameFormatOverride string        `yaml:"filename_format_override"`
	Renames                []RenameItem  `yaml:"renames"`
	InstallNote            string        `yaml:"install_note"`
	RemoveNote             string        `yaml:"remove_note"`
}

type OsArchPair struct {
	Os   string `yaml:"os"`
	Arch string `yaml:"arch"`
}

type PkgConfig struct {
	Title       string `yaml:"-"`
	Tagline     string `yaml:"tagline"`
	About       string `yaml:"about"`
	InstallNote string `yaml:"install_note"`
	RemoveNote  string `yaml:"remove_note"`

	InfoUrl         string `yaml:"info_url"`
	ReleasesUrl     string `yaml:"releases_url"`
	BaseDownloadUrl string `yaml:"base_download_url"`
	GitUser         string `yaml:"git_user"`
	GitRepo         string `yaml:"git_repo"`
	GiteaURL        string `yaml:"gitea_url"`
	SourceUrl       string `yaml:"source_url"`

	FilenameFormat   string `yaml:"filename_format"`
	VersionFormat    string `yaml:"version_format"`
	LatestStrategy   string `yaml:"latest_strategy"`
	ForceLatest      bool   `yaml:"force_latest"`
	AllowPrerelease  bool   `yaml:"allow_prerelease"`
	ArchLinuxPkgName string `yaml:"arch_linux_pkg_name"`

	OsMap   map[string]OsInfo `yaml:"os_map"`
	ArchMap map[string]string `yaml:"arch_map"`
	Ignore  []OsArchPair      `yaml:"ignore"`
}

func (pkgConf *PkgConfig) InstallNotes() string {
	var installNotes string

	pkgOS := GOOStoPkgOs[utils.GOOS]
	note := pkgConf.InstallNote
	osNote := pkgConf.OsMap[pkgOS].InstallNote
	if note != "" || osNote != "" {
		installNotes += color.BlueString("== %s\n", pkgConf.Title)
	}
	if note != "" {
		installNotes += color.YellowString(note) + "\n"
	}
	if osNote != "" {
		installNotes += color.YellowString(osNote) + "\n"
	}
	return installNotes
}

func (pkgConf *PkgConfig) RemoveNotes() string {
	var removeNotes string

	pkgOS := GOOStoPkgOs[utils.GOOS]
	note := pkgConf.RemoveNote
	osNote := pkgConf.OsMap[pkgOS].RemoveNote
	if note != "" || osNote != "" {
		removeNotes += color.BlueString("== %s\n", pkgConf.Title)
	}
	if note != "" {
		removeNotes += color.YellowString(note) + "\n"
	}
	if osNote != "" {
		removeNotes += color.YellowString(osNote) + "\n"
	}
	return removeNotes
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
	osStr, exists := GOOStoPkgOs[utils.GOOS]
	if !exists {
		return []string{}, fmt.Errorf("unsupported OS")
	}
	osInfo, exists := pkgConf.OsMap[osStr]
	if !exists {
		return []string{}, fmt.Errorf("package does not support this OS")
	}
	if osInfo.IsRawBinary {
		return []string{pkgConf.Title}, nil
	}
	if len(osInfo.BinPaths.Values) == 0 {
		osInfo.BinPaths = SingleOrMulti{[]string{""}}
	}
	return osInfo.BinPaths.Values, nil
}

func (pkgConf *PkgConfig) GetRenames() ([]RenameItem, error) {
	osStr, exists := GOOStoPkgOs[utils.GOOS]
	if !exists {
		return []RenameItem{}, fmt.Errorf("unsupported OS")
	}
	osInfo, exists := pkgConf.OsMap[osStr]
	if !exists {
		return []RenameItem{}, fmt.Errorf("package does not support this OS")
	}
	return osInfo.Renames, nil
}

// Check using file.
// If UsingFile doesn't exist, it is not using anything
func CheckUsing(pkg string) (*string, error) {
	usingPath := filepath.Join(utils.WebmanPkgDir, pkg, utils.UsingFileName)
	usingContent, err := os.ReadFile(usingPath)
	if err != nil {
		return nil, nil
	}
	var usingInfo UsingInfo
	if err = yaml.Unmarshal(usingContent, &usingInfo); err != nil {
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
	usingPath := filepath.Join(utils.WebmanPkgDir, pkg, utils.UsingFileName)
	if err := os.WriteFile(usingPath, data, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func RemoveUsing(pkg string) error {
	usingPath := filepath.Join(utils.WebmanPkgDir, pkg, utils.UsingFileName)
	if err := os.Remove(usingPath); err != nil {
		return err
	}
	return nil
}

func ParsePkgConfig(name string, r io.Reader) (*PkgConfig, error) {
	dat, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var pkgConf PkgConfig
	if err = yaml.Unmarshal(dat, &pkgConf); err != nil {
		return nil, fmt.Errorf("unable to parse package recipe for %s: %v", name, err)
	}
	pkgConf.Title = name

	pkgConf.BaseDownloadUrl = strings.ReplaceAll(pkgConf.BaseDownloadUrl, "[GIT_USER]", pkgConf.GitUser)
	pkgConf.BaseDownloadUrl = strings.ReplaceAll(pkgConf.BaseDownloadUrl, "[GIT_REPO]", pkgConf.GitRepo)

	pkgConf.InfoUrl = strings.ReplaceAll(pkgConf.InfoUrl, "[GIT_USER]", pkgConf.GitUser)
	pkgConf.InfoUrl = strings.ReplaceAll(pkgConf.InfoUrl, "[GIT_REPO]", pkgConf.GitRepo)

	pkgConf.ReleasesUrl = strings.ReplaceAll(pkgConf.InfoUrl, "[GIT_USER]", pkgConf.GitUser)
	pkgConf.ReleasesUrl = strings.ReplaceAll(pkgConf.InfoUrl, "[GIT_REPO]", pkgConf.GitRepo)

	pkgConf.SourceUrl = strings.ReplaceAll(pkgConf.SourceUrl, "[GIT_USER]", pkgConf.GitUser)
	pkgConf.SourceUrl = strings.ReplaceAll(pkgConf.SourceUrl, "[GIT_REPO]", pkgConf.GitRepo)

	pkgConf.GiteaURL = strings.TrimRight(pkgConf.GiteaURL, "/")

	return &pkgConf, nil
}

func ParsePkgConfigLocal(pkgRepos []*config.PkgRepo, pkg string) (*PkgConfig, error) {
	var pkgConfPath string
	for _, pkgRepo := range pkgRepos {
		pkgPath := filepath.Join(pkgRepo.Path(), "pkgs", pkg+utils.PkgRecipeExt)
		_, err := os.Stat(pkgPath)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return nil, err
			}
			continue
		}
		pkgConfPath = pkgPath
		break
	}
	if pkgConfPath == "" {
		return nil, fmt.Errorf("no package recipe exists for %s", pkg)
	}

	fi, err := os.Open(pkgConfPath)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	return ParsePkgConfig(pkg, fi)
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
	case "gitea-release":
		rel, err := getLatestGiteaReleaseTag(pkgConf.GiteaURL, pkgConf.GitUser, pkgConf.GitRepo, pkgConf.AllowPrerelease)
		if err != nil {
			return nil, err
		}
		version = rel.TagName
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

func (pkgConf *PkgConfig) GetAssetStemExtUrl(version string) (*string, *string, *string, error) {
	pkgOs, exists := GOOStoPkgOs[utils.GOOS]
	if !exists {
		return nil, nil, nil, fmt.Errorf("unsupported operating system")
	}
	osInf, exists := pkgConf.OsMap[pkgOs]
	if !exists {
		return nil, nil, nil, fmt.Errorf("package has no binary for operating system: %s", pkgOs)
	}
	archStr, exists := pkgConf.ArchMap[utils.GOARCH]
	if !exists {
		return nil, nil, nil, fmt.Errorf("package has no binary for architecture: %s", utils.GOARCH)
	}
	baseUrl := pkgConf.BaseDownloadUrl
	baseUrl = strings.ReplaceAll(baseUrl, "[VER]", version)
	baseUrl = strings.ReplaceAll(baseUrl, "[OS]", osInf.Name)
	baseUrl = strings.ReplaceAll(baseUrl, "[ARCH]", archStr)
	baseUrl = strings.ReplaceAll(baseUrl, "[EXT]", osInf.Ext)

	fileStem := pkgConf.FilenameFormat
	if osInf.FilenameFormatOverride != "" {
		fileStem = osInf.FilenameFormatOverride
	}
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

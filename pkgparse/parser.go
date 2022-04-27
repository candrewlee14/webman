package pkgparse

import (
    "fmt"
	"os"
	"path/filepath"
	"github.com/goccy/go-yaml"
    "net/http"
    "io/ioutil"
    "encoding/json"
    "runtime"
    "strings"
    "io"
)

type OsInfo struct {
    Name string
    Ext string
}

type PkgConfig struct {
    InfoUrl string `json:"info_url"`
    ReleasesUrl string `json:"releases_url"`
    BaseDownloadUrl string `json:"base_download_url"`
    GitUser string `json:"git_user"`
    GitRepo string `json:"git_repo"`
    SourceUrl string `json:"source_url"`

    FilenameFormat string `json:"filename_format"`
    VersionType string `json:"version_type"`
    LatestStrategy string `json:"latest_strategy"`
    ArchLinuxPkgName string `json:"arch_linux_pkg_name"`

    OsMap map[string]OsInfo `json:"os_map"`
    ArchMap map[string]string `json:"arch_map"`
}

var pkgConfDir string = "./pkgs"

func ParsePkgConfig(pkg string) PkgConfig {
    curDir, err := os.Getwd()
    if err != nil {
        panic(err)
    }
    pkgConfPath := filepath.Join(curDir, "/pkgs/" + pkg + ".yaml") 
    dat, err := os.ReadFile(pkgConfPath)
    if err != nil {
        panic(err)
    }
    var pkgConf PkgConfig
    err = yaml.Unmarshal(dat, &pkgConf)
    if err != nil {
        panic(err)
    }
    return pkgConf
}

var goOsToPkgOs = map[string]string {
    "darwin" : "macos",
    "windows" : "win",
    "linux" : "linux",
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

func GetAsset(url string, w io.Writer) {
    r, err := http.Get(url)
    if err != nil {
        panic(err)
    }
    defer r.Body.Close()
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        panic(err)
    }
    var rInfo ReleaseInfo;
    err = json.Unmarshal(body, &rInfo)
    if err != nil {
        panic(err)
    }
    if len(rInfo.Url) == 0 {
        fmt.Println("Release not found");
    } else {
        fmt.Println(rInfo.Url);
        fmt.Println(len(rInfo.Assets))
        fmt.Println(rInfo.Assets[0]);
        fmt.Println("No problems!");
    }
}

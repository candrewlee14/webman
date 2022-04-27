package pkgparse
import (
    "io/ioutil"
    "net/http"
    "fmt"
    "encoding/json"
)

type ReleaseInfo struct {
   Url string
   Assets []AssetInfo
   TagName string `json:"tag_name"`
   Date string `json:"published_at"`
}

type AssetInfo struct {
    Name string
    Size uint32
    BrowserDownloadUrl string `json:"browser_download_url"`
}

// func parseAssetName(name string) map[string]string {
//     var keys = map[string]string{
//         "VERSION": `(?P<version>\d+\.\d+\.\d+)`,
//         "ARCH": `(?P<arch>[^-]+)`,
//         "VENDOR": `(?P<os>\w+)`,
//         "OS": `(?P<os>\w+)`,
//         "ABI": `(?P<abi>\w+)`,
//         "EXT": `(?P<ext>[^ ]+)`,
//     }
//     myExp := regexp.MustCompile(
//         `(\w+)-` + semVer + "-" + 
//         arch + "-" + vendorOs + optAbi + `\.` + fileExt)
//     match := myExp.FindStringSubmatch(name)
//     result := make(map[string]string)
//     for i, name := range myExp.SubexpNames() {
//         if i != 0 && name != "" {
//             result[name] = match[i]
//         }
//     }
//     return result
// }

type ReleaseTagInfo struct {
   TagName string `json:"tag_name"`
   Date string `json:"published_at"`
}

func getLatestGithubReleaseTag(user string, repo string) ReleaseTagInfo {
    url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", user, repo)
    r, err := http.Get(url)
    if err != nil {
        panic(err)
    }
    defer r.Body.Close()
    if !(r.StatusCode >= 200 && r.StatusCode < 300) {
        panic("Bad HTTP Response: " + r.Status)
    }
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        panic(err)
    }
    var releases []ReleaseTagInfo;
    err = json.Unmarshal(body, &releases)
    if len(releases) == 0 {
        panic(fmt.Sprintf("Expected at least one release listed at %s, unable to resolve latest", url))
    }
    return releases[0]
}


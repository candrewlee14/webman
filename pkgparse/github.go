package pkgparse

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/mholt/archiver/v3"

	"github.com/candrewlee14/webman/utils"

	"gopkg.in/yaml.v3"
)

type ReleaseInfo struct {
	Url     string
	Assets  []AssetInfo
	TagName string `json:"tag_name"`
	Date    string `json:"published_at"`
}

type AssetInfo struct {
	Name               string
	Size               uint32
	BrowserDownloadUrl string `json:"browser_download_url"`
}

type ReleaseTagInfo struct {
	TagName    string `json:"tag_name"`
	Date       string `json:"published_at"`
	Prerelease bool
	Draft      bool
}

func getLatestGithubReleaseTag(user string, repo string, allowPrerelease bool) (*ReleaseTagInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", user, repo)
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	if !(r.StatusCode >= 200 && r.StatusCode < 300) {
		return nil, fmt.Errorf("bad HTTP Response: %s", r.Status)
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var releases []ReleaseTagInfo
	if err = json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("github releases JSON response not in expected format")
	}
	if len(releases) == 0 {
		return nil, fmt.Errorf("expected at least one release listed at %s, unable to resolve latest", url)
	}
	for _, release := range releases {
		if (allowPrerelease || !release.Prerelease) && !release.Draft {
			return &release, nil
		}
	}
	return nil, fmt.Errorf("found no stable releases for %s/%s", user, repo)
}

type GithubDir struct {
	Name        string
	DownloadUrl string `yaml:"download_url"`
}

type RefreshFile struct {
	LastUpdated *time.Time
}

func ShouldRefreshRecipes() (bool, error) {
	refreshFileDir := filepath.Join(utils.WebmanRecipeDir, utils.RefreshFileName)
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
	if refreshFile.LastUpdated == nil {
		return true, nil
	}
	timeSince := time.Since(*refreshFile.LastUpdated)
	if timeSince > (time.Hour * 6) {
		return true, nil
	}
	return false, nil
}

func RefreshRecipes() error {
	if err := os.RemoveAll(utils.WebmanRecipeDir); err != nil {
		return err
	}
	url := "https://api.github.com/repos/candrewlee14/webman-pkgs/zipball/main"
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if !(r.StatusCode >= 200 && r.StatusCode < 300) {
		return fmt.Errorf("Bad HTTP Response: " + r.Status)
	}
	if err = os.RemoveAll(utils.WebmanRecipeDir); err != nil {
		return err
	}
	if err = os.MkdirAll(utils.WebmanTmpDir, os.ModePerm); err != nil {
		return err
	}
	tmpZipFile, err := os.CreateTemp(utils.WebmanTmpDir, "recipes-*.zip")
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
	if err = os.Rename(innerTmpFolder, utils.WebmanRecipeDir); err != nil {
		return err
	}
	refreshFilePath := filepath.Join(utils.WebmanRecipeDir, utils.RefreshFileName)
	curTime := time.Now()
	data, err := yaml.Marshal(RefreshFile{&curTime})
	if err != nil {
		return nil
	}
	if err = os.WriteFile(refreshFilePath, data, os.ModePerm); err != nil {
		return err
	}
	return nil
}

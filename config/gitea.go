package config

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/candrewlee14/webman/utils"

	"github.com/mholt/archiver/v3"
	"gopkg.in/yaml.v3"
)

func (p PkgRepo) giteaRefresh() error {
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s/archive/%s.tar.gz", p.GiteaURL, p.User, p.Repo, p.Branch)
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if !(r.StatusCode >= 200 && r.StatusCode < 300) {
		return fmt.Errorf("Bad HTTP Response: " + r.Status)
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
	refreshFilePath := filepath.Join(p.Path(), utils.RefreshFileName)
	data, err := yaml.Marshal(RefreshFile{time.Now()})
	if err != nil {
		return nil
	}
	if err = os.WriteFile(refreshFilePath, data, os.ModePerm); err != nil {
		return err
	}
	return nil
}

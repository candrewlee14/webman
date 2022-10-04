package pkgparse

import (
	"os"
	"path/filepath"

	"github.com/candrewlee14/webman/utils"

	"gopkg.in/yaml.v3"
)

// UsingInfo is which version of a package is being used
type UsingInfo struct {
	Using string `yaml:"using"`
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

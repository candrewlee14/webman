package pkgparse

import (
	"fmt"
	"os"
	"path/filepath"
	"webman/utils"

	"github.com/fatih/color"
	"github.com/go-yaml/yaml"
)

type PkgGroupConfig struct {
	Title   string
	Tagline string
	About   string

	InfoUrl  string   `yaml:"info_url"`
	Packages []string `yaml:"packages"`
}

func ParseGroupConfig(group string) *PkgGroupConfig {
	groupPath := filepath.Join(utils.WebmanRecipeDir, "groups", group+".yaml")
	fmt.Println(groupPath)
	if _, err := os.Stat(groupPath); err != nil {
		if os.IsNotExist(err) {
			color.Red("No package group named %s", color.YellowString(group))
			os.Exit(1)
		}
		color.Red("Error accessing package group: %v", err)
		os.Exit(1)
	}
	var groupConf PkgGroupConfig
	data, err := os.ReadFile(groupPath)
	if err != nil {
		color.Red("Failed to read package group file: %v", err)
		os.Exit(1)
	}
	if err = yaml.UnmarshalStrict(data, &groupConf); err != nil {
		color.Red("Invalid format for package group: %v", err)
		os.Exit(1)
	}
	if len(groupConf.Packages) == 0 {
		color.Red("No packages in package group %s", color.YellowString(group))
		os.Exit(1)
	}
	return &groupConf
}

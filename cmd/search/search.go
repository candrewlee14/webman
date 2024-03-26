package search

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/candrewlee14/webman/cmd/add"
	"github.com/candrewlee14/webman/config"
	"github.com/candrewlee14/webman/multiline"
	"github.com/candrewlee14/webman/pkgparse"
	"github.com/candrewlee14/webman/utils"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
)

var doRefresh bool

var SearchCmd = &cobra.Command{
	Use:   "search",
	Short: "search for a package",
	Long: `
The "search" subcommand starts an interactive window to find and display info about a package`,
	Example: `webman search`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return cmd.Help()
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		// if local recipe flag is not set
		if utils.RecipeDirFlag == "" {
			// only refresh if not using local
			for _, pkgRepo := range cfg.PkgRepos {
				shouldRefresh, err := pkgRepo.ShouldRefreshRecipes(cfg.RefreshInterval)
				if err != nil {
					return err
				}
				if shouldRefresh || doRefresh {
					color.HiBlue("Refreshing package recipes for %q...", pkgRepo.Name)
					if err = pkgRepo.RefreshRecipes(); err != nil {
						color.Red("%v", err)
					}
				}
			}
		}
		pkgInfos := make([]*pkgparse.PkgConfig, 0)
		for _, pkgRepo := range cfg.PkgRepos {
			files, err := os.ReadDir(pkgRepo.PackagePath())
			if err != nil {
				return err
			}
			for _, file := range files {
				pkg := strings.Split(file.Name(), utils.PkgRecipeExt)[0]
				pkgInfo, err := pkgparse.ParsePkgConfigPath(pkgRepo.Path(), pkg)
				if err != nil {
					return err
				}
				pkgInfos = append(pkgInfos, pkgInfo)
			}
		}
		sort.Slice(pkgInfos, func(i, j int) bool {
			return pkgInfos[i].Title < pkgInfos[j].Title
		})

		installed := utils.InstalledPackages()
		installedSet := make(map[string]struct{})
		for _, i := range installed {
			installedSet[i] = struct{}{}
		}
		idx, err := fuzzyfinder.Find(
			pkgInfos,
			func(i int) string {
				pre := "   "
				if _, ok := installedSet[pkgInfos[i].Title]; ok {
					pre = "✅ "
				}
				return pre + pkgInfos[i].Title + " - " + pkgInfos[i].Tagline
			},
			fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
				if i == -1 {
					return ""
				}
				preview := fmt.Sprintf("%s: %s\n\n%s:\n %s\n\n%s:\n%s",
					"📦 Title",
					pkgInfos[i].Title,
					"💾 Tagline",
					pkgInfos[i].Tagline,
					"📄 About",
					pkgInfos[i].About,
				)
				notes := pkgInfos[i].InstallNotes()
				if notes != "" {
					preview += fmt.Sprintf("\n\n%s:\n %s",
						"📝 Notes",
						pkgInfos[i].InstallNotes(),
					)
				}
				return wrapText(preview, w)
			}))
		if err != nil {
			color.HiBlack("No package selected.")
			return nil
		}
		pkgName := pkgInfos[idx].Title
		prompt := &survey.Confirm{
			Message: "Would you like to install the latest version of " + color.CyanString(pkgName) + "?",
		}
		shouldInstall := false
		if err := survey.AskOne(prompt, &shouldInstall); err != nil || !shouldInstall {
			color.HiBlack("No package selected.")
			return nil
		}
		var wg sync.WaitGroup
		ml := multiline.New(1, os.Stdout)
		wg.Add(1)
		pkg := add.InstallPkg(cfg.PkgRepos, pkgName, 0, 1, &wg, &ml, false, false)
		if pkg == nil {
			return errors.New("failed to install pkg")
		}
		fmt.Print(pkg.PkgConf.InstallNotes())
		return nil
	},
}

func init() {
	SearchCmd.Flags().BoolVar(&doRefresh, "refresh", false, "force refresh of package recipes")
}

func wrapText(text string, width int) string {
	prevI := -1
	var buf strings.Builder
	for i, ch := range text {
		if ch == '\n' || i-prevI > (width/2-6) {
			io.WriteString(&buf, strings.TrimSpace(text[prevI+1:i+1])+"\n")
			prevI = i
		}
	}
	return buf.String()
}

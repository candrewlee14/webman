package search

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/candrewlee14/webman/cmd/group/add"
	"github.com/candrewlee14/webman/config"
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
	Short: "search for a group",
	Long: `
The "search" subcommand starts an interactive window to find and display info about a group`,
	Example: `webman group search`,
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

		groupInfos := make([]*pkgparse.PkgGroupConfig, 0)
		for _, pkgRepo := range cfg.PkgRepos {
			files, err := os.ReadDir(pkgRepo.GroupPath())
			if err != nil {
				return err
			}
			for _, file := range files {
				group := strings.Split(file.Name(), utils.GroupRecipeExt)[0]
				groupInfo, err := pkgparse.ParseGroupConfigInRepo(pkgRepo, group)
				if err != nil {
					return err
				}
				groupInfos = append(groupInfos, groupInfo)
			}
		}
		sort.Slice(groupInfos, func(i, j int) bool {
			return groupInfos[i].Title < groupInfos[j].Title
		})

		idx, err := fuzzyfinder.Find(
			groupInfos,
			func(i int) string {
				return groupInfos[i].Title + " - " + groupInfos[i].Tagline
			},
			fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
				if i == -1 {
					return ""
				}
				preview := fmt.Sprintf("%s: %s\n\n%s:\n %s\n\n%s:\n%s\n\n%s:\n%s\n",
					"📚 Title",
					groupInfos[i].Title,
					"💾 Tagline",
					groupInfos[i].Tagline,
					"🛈 Description",
					groupInfos[i].Description,
					"🗐 Package List",
					strings.Join(groupInfos[i].Packages, ", "),
				)
				return wrapText(preview, w)
			}))
		if err != nil {
			color.HiBlack("No group selected.")
			return nil
		}
		groupName := groupInfos[idx].Title
		prompt := &survey.Confirm{
			Message: "Would you like to install the latest version of " + color.CyanString(groupName) + "?",
		}
		shouldInstall := false
		if err := survey.AskOne(prompt, &shouldInstall); err != nil || !shouldInstall {
			color.HiBlack("No group selected.")
			return nil
		}
		return add.InstallGroup(cfg, groupName)
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

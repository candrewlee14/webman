package search

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/candrewlee14/webman/cmd/add"
	"github.com/candrewlee14/webman/multiline"
	"github.com/candrewlee14/webman/pkgparse"
	"github.com/candrewlee14/webman/utils"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
)

var SearchCmd = &cobra.Command{
	Use:   "search",
	Short: "search for a package",
	Long: `
The "search" subcommand starts an interactive window to find and display info about a package`,
	Example: `webman search`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			cmd.Help()
			return nil
		}
		utils.Init()
		files, err := os.ReadDir(filepath.Join(utils.WebmanRecipeDir, "pkgs"))
		if err != nil {
			panic(err)
		}
		pkgInfos := make([]*pkgparse.PkgInfo, len(files))
		for i, file := range files {
			pkg := strings.Split(file.Name(), utils.PkgRecipeExt)[0]
			pkgInfo, err := pkgparse.ParsePkgInfo(pkg)
			if err != nil {
				panic(err)
			}
			pkgInfos[i] = pkgInfo
		}
		idx, err := fuzzyfinder.Find(
			pkgInfos,
			func(i int) string {
				return pkgInfos[i].Title + " - " + pkgInfos[i].Tagline
			},
			fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
				if i == -1 {
					return ""
				}
				return wrapText(fmt.Sprintf("%s: %s\n\n%s:\n %s\n\n%s:\n%s",
					"📦 Title",
					pkgInfos[i].Title,
					"💾 Tagline",
					pkgInfos[i].Tagline,
					"📄 About",
					pkgInfos[i].About), w)

			}))
		if err != nil {
			color.HiBlack("No package selected.")
			return nil
		}
		pkg := pkgInfos[idx].Title
		prompt := &survey.Confirm{
			Message: "Would you like to install the latest version of " + color.CyanString(pkg) + "?",
		}
		shouldInstall := false
		if err := survey.AskOne(prompt, &shouldInstall); err != nil || !shouldInstall {
			color.HiBlack("No package selected.")
			return nil
		}
		var wg sync.WaitGroup
		ml := multiline.New(1, os.Stdout)
		wg.Add(1)
		if !add.InstallPkg(pkg, 0, 1, &wg, &ml) {
			return errors.New("failed to install pkg")
		}
		return nil
	},
}

func init() {

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// removeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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

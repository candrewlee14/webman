package remove

import (
	"errors"
	"os"

	"github.com/candrewlee14/webman/config"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var RemoveCmd = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"delete"},
	Short:   "remove a package repository",
	Long: `

The "config remove" subcommand allows you to remove a package repository.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(cfg.PkgRepos) == 1 {
			return errors.New("cannot remove only package repository")
		}

		opts := make([]string, 0, len(cfg.PkgRepos))
		for _, pkgRepo := range cfg.PkgRepos {
			opts = append(opts, pkgRepo.Name)
		}

		q := &survey.Select{
			Message: "Repository to remove",
			Options: opts,
		}

		var pkg string
		if err := survey.AskOne(q, &pkg, survey.WithValidator(survey.Required)); err != nil {
			return err
		}

		var remove *config.PkgRepo
		for idx, pkgRepo := range cfg.PkgRepos {
			if pkgRepo.Name == pkg {
				remove = pkgRepo
				cfg.PkgRepos = append(cfg.PkgRepos[:idx], cfg.PkgRepos[idx+1:]...)
				break
			}
		}

		if remove != nil {
			if err := os.RemoveAll(remove.Path()); err != nil {
				return err
			}
		}

		if err := cfg.Save(); err != nil {
			return err
		}

		color.HiGreen("Repository %q successfully removed", pkg)
		return nil
	},
}

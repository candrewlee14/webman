package add

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
	"github.com/candrewlee14/webman/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "add a package repository",
	Long: `

The "config add" subcommand allows you to add a package repository.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		var repo PkgRepo
		qs := []*survey.Question{
			{
				Name: "name",
				Prompt: &survey.Input{
					Message: "Repository name",
				},
				Validate: survey.Required,
			},
			{
				Name: "type",
				Prompt: &survey.Select{
					Message: "Repository type",
					Options: []string{"gitea", "github"},
				},
				Validate: survey.Required,
			},
			{
				Name: "user",
				Prompt: &survey.Input{
					Message: "Git user name",
				},
				Validate: survey.Required,
			},
			{
				Name: "repo",
				Prompt: &survey.Input{
					Message: "Git repository name",
				},
				Validate: survey.Required,
			},
			{
				Name: "branch",
				Prompt: &survey.Input{
					Message: "Git branch name",
					Default: "main",
				},
			},
		}

		if err := survey.Ask(qs, &repo); err != nil {
			return err
		}

		if repo.Type == config.PkgRepoTypeGitea {
			q := &survey.Input{
				Message: "Gitea URL",
			}
			if err := survey.AskOne(q, &repo.GiteaURL, survey.WithValidator(survey.Required)); err != nil {
				return err
			}
		}

		p := config.PkgRepo(repo)

		if err := p.RefreshRecipes(); err != nil {
			return err
		}

		cfg.PkgRepos = append(cfg.PkgRepos, &p)

		if err := cfg.Save(); err != nil {
			return err
		}

		color.HiGreen("Repository %q successfully added", repo.Name)
		return nil
	},
}

// PkgRepo overrides config.PkgRepo in order to implement WriteAnswer without polluting the config package
// with survey details
type PkgRepo config.PkgRepo

func (p *PkgRepo) WriteAnswer(field string, value any) error {
	switch field {
	case "name":
		p.Name = fmt.Sprint(value)
	case "type":
		p.Type = config.PkgRepoType(value.(core.OptionAnswer).Value)
	case "user":
		p.User = fmt.Sprint(value)
	case "repo":
		p.Repo = fmt.Sprint(value)
	case "branch":
		p.Branch = fmt.Sprint(value)
	default:
		return errors.New("unknown field")
	}
	return nil
}

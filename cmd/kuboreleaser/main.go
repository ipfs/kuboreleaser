package main

import (
	"log"
	"os"

	"github.com/ipfs/kuboreleaser/actions"
	"github.com/ipfs/kuboreleaser/git"
	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/matrix"
	"github.com/ipfs/kuboreleaser/util"
	"github.com/urfave/cli/v2"
)

func Execute(action actions.IAction, c *cli.Context) error {

	if !c.Bool("skip-check-before") {
		if err := action.Check(); err != nil {
			return err
		}
	}

	if !c.Bool("skip-run") {
		if err := action.Run(); err != nil {
			return err
		}
	}

	if !c.Bool("skip-check-after") {
		if err := action.Check(); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	app := &cli.App{
		Name:  "kuboreleaser",
		Usage: "Kubo Release CLI",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "skip-check-before",
				Aliases: []string{"scb"},
				Usage:   "skip the check of the command before the run",
			}, &cli.BoolFlag{
				Name:    "skip-run",
				Aliases: []string{"sr"},
				Usage:   "skip the run of the command",
			}, &cli.BoolFlag{
				Name:    "skip-check-after",
				Aliases: []string{"sca"},
				Usage:   "skip the check of the command after the run",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "release",
				Usage: "Release Kubo",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "version",
						Aliases: []string{"v"},
						Usage:   "Kubo version to release",
					},
				},
				Before: func(c *cli.Context) error {
					git, err := git.NewClient()
					if err != nil {
						return err
					}

					github, err := github.NewClient()
					if err != nil {
						return err
					}

					matrix, err := matrix.NewClient()
					if err != nil {
						return err
					}

					version, err := util.NewVersion(c.String("version"))
					if err != nil {
						return err
					}

					c.App.Metadata["git"] = git
					c.App.Metadata["github"] = github
					c.App.Metadata["matrix"] = matrix
					c.App.Metadata["version"] = version

					return nil
				},
				Subcommands: []*cli.Command{
					{
						Name:  "cut-branch",
						Usage: "Cut a new branch",
						Action: func(c *cli.Context) error {
							git := c.App.Metadata["git"].(*git.Client)
							github := c.App.Metadata["github"].(*github.Client)
							version := c.App.Metadata["version"].(*util.Version)

							action, err := actions.NewCutBranch(git, github, version)
							if err != nil {
								return err
							}

							return Execute(action, c)
						},
					},
					{
						Name:  "notify-bifrost",
						Usage: "Notify Bifrost of the new release",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name: "date",
								Aliases: []string{"d"},
								Usage: "Date of the release",
							},
						},
						Action: func(c *cli.Context) error {
							github := c.App.Metadata["github"].(*github.Client)
							version := c.App.Metadata["version"].(*util.Version)

							action, err := actions.NewNotifyBifrost(github, version, c.Timestamp("date"))
							if err != nil {
								return err
							}

							return Execute(action, c)
						},
					},
					{
						Name:  "promote",
						Usage: "Promote the release",
						Action: func(c *cli.Context) error {
							github := c.App.Metadata["github"].(*github.Client)
							matrix := c.App.Metadata["matrix"].(*matrix.Client)
							version := c.App.Metadata["version"].(*util.Version)

							action, err := actions.NewPromote(github, matrix, version)
							if err != nil {
								return err
							}

							return Execute(action, c)
						},
					},
					{
						Name:  "publish-to-distributions",
						Usage: "Publish the release to distributions",
						Action: func(c *cli.Context) error {
							git := c.App.Metadata["git"].(*git.Client)
							github := c.App.Metadata["github"].(*github.Client)
							version := c.App.Metadata["version"].(*util.Version)

							action, err := actions.NewPublishToDistributions(git, github, version)
							if err != nil {
								return err
							}

							return Execute(action, c)
						},
					},
					{
						Name:  "publish-to-github",
						Usage: "Publish the release to GitHub",
						Action: func(c *cli.Context) error {
							github := c.App.Metadata["github"].(*github.Client)
							version := c.App.Metadata["version"].(*util.Version)

							action, err := actions.NewPublishToGitHub(github, version)
							if err != nil {
								return err
							}

							return Execute(action, c)
						},
					},
					{
						Name:  "publish-to-npm",
						Usage: "Publish the release to npm",
						Action: func(c *cli.Context) error {
							github := c.App.Metadata["github"].(*github.Client)
							version := c.App.Metadata["version"].(*util.Version)

							action, err := actions.NewPublishToNPM(github, version)
							if err != nil {
								return err
							}

							return Execute(action, c)
						},
					},
					{
						Name:  "tag",
						Usage: "Tag the release",
						Action: func(c *cli.Context) error {
							git := c.App.Metadata["git"].(*git.Client)
							github := c.App.Metadata["github"].(*github.Client)
							version := c.App.Metadata["version"].(*util.Version)

							action, err := actions.NewTag(git, github, version)
							if err != nil {
								return err
							}

							return Execute(action, c)
						},
					},
					{
						Name:  "test-ipfs-companion",
						Usage: "Test the release with ipfs-companion",
						Action: func(c *cli.Context) error {
							github := c.App.Metadata["github"].(*github.Client)
							version := c.App.Metadata["version"].(*util.Version)

							action, err := actions.NewTestIPFSCompanion(github, version)
							if err != nil {
								return err
							}

							return Execute(action, c)
						},
					},
					{
						Name:  "update-ipfs-desktop",
						Usage: "Update the release in ipfs-desktop",
						Action: func(c *cli.Context) error {
							git := c.App.Metadata["git"].(*git.Client)
							github := c.App.Metadata["github"].(*github.Client)
							version := c.App.Metadata["version"].(*util.Version)

							action, err := actions.NewUpdateIPFSDesktop(git, github, version)
							if err != nil {
								return err
							}

							return Execute(action, c)
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

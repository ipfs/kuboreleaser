package main

import (
	"errors"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ipfs/kuboreleaser/actions"
	"github.com/ipfs/kuboreleaser/git"
	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/matrix"
	"github.com/ipfs/kuboreleaser/util"
	"github.com/urfave/cli/v2"
)

func Execute(action actions.IAction, c *cli.Context) error {
	if !c.Bool("skip-check-before") {
		log.Info("Checking the status of the action...")
		err := action.Check()
		if err != nil {
			if !errors.Is(err, actions.ErrIncomplete) {
				return err
			} else {
				log.Info("The action is not complete yet, continuing...", err)
			}
		} else {
			log.Info("Action already completed")
			return nil
		}
	} else {
		log.Info("Skipping the check before running the action")
	}

	if !c.Bool("skip-run") {
		log.Info("Running the action...")
		err := action.Run()
		if err != nil {
			return err
		}
	} else {
		log.Info("Skipping the run of the action")
	}

	if !c.Bool("skip-check-after") {
		duration := time.Second * 10
		for {
			log.Info("Sleeping for ", duration, "...")
			time.Sleep(duration)
			if duration < time.Minute*10 {
				duration = duration * 2
			}

			log.Info("Checking the status of the action...")
			err := action.Check()
			if err != nil {
				if errors.Is(err, actions.ErrInProgress) && !c.Bool("skip-wait") {
					log.Info("The action is still in progress, continuing...", err)
					continue
				}
				return err
			}
			log.Info("Action completed")
			return nil
		}
	} else {
		log.Info("Skipping the check after running the action")
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
			}, &cli.BoolFlag{
				Name:    "skip-wait",
				Aliases: []string{"sw"},
				Usage:   "skip the wait for the command to complete after the run",
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
					log.Debug("Initializing git client...")
					git, err := git.NewClient()
					if err != nil {
						return err
					}

					log.Debug("Initializing GitHub client...")
					github, err := github.NewClient()
					if err != nil {
						return err
					}

					log.Debug("Initializing Matrix client...")
					matrix, err := matrix.NewClient()
					if err != nil {
						return err
					}

					log.Debug("Initializing version...")
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
						Name:  "prepare-branch",
						Usage: "Prepare a branch for the release",
						Action: func(c *cli.Context) error {
							git := c.App.Metadata["git"].(*git.Client)
							github := c.App.Metadata["github"].(*github.Client)
							version := c.App.Metadata["version"].(*util.Version)

							action := &actions.PrepareBranch{
								Git:     git,
								GitHub:  github,
								Version: version,
							}

							return Execute(action, c)
						},
					},
					{
						Name:  "notify-bifrost",
						Usage: "Notify Bifrost of the new release",
						Flags: []cli.Flag{
							&cli.TimestampFlag{
								Name:    "date",
								Aliases: []string{"d"},
								Usage:   "Date of the release",
								Layout:  "2006-01-02",
							},
						},
						Action: func(c *cli.Context) error {
							github := c.App.Metadata["github"].(*github.Client)
							version := c.App.Metadata["version"].(*util.Version)

							action := &actions.NotifyBifrost{
								GitHub:  github,
								Version: version,
								Date:    c.Timestamp("date"),
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

							action := &actions.Promote{
								GitHub:  github,
								Matrix:  matrix,
								Version: version,
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

							action := &actions.PublishToDistributions{
								Git:     git,
								GitHub:  github,
								Version: version,
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

							action := &actions.PublishToGitHub{
								GitHub:  github,
								Version: version,
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

							action := &actions.PublishToNPM{
								GitHub:  github,
								Version: version,
							}

							return Execute(action, c)
						},
					},
					{
						Name:  "publish-to-dockerhub",
						Usage: "Publish the release to DockerHub",
						Action: func(c *cli.Context) error {
							github := c.App.Metadata["github"].(*github.Client)
							version := c.App.Metadata["version"].(*util.Version)

							action := &actions.PublishToDockerHub{
								GitHub:  github,
								Version: version,
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

							action := &actions.Tag{
								Git:     git,
								GitHub:  github,
								Version: version,
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

							action := &actions.TestIPFSCompanion{
								GitHub:  github,
								Version: version,
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

							action := &actions.UpdateIPFSDesktop{
								Git:     git,
								GitHub:  github,
								Version: version,
							}

							return Execute(action, c)
						},
					},
					{
						Name:  "update-interop",
						Usage: "Update the release in interop",
						Action: func(c *cli.Context) error {
							git := c.App.Metadata["git"].(*git.Client)
							github := c.App.Metadata["github"].(*github.Client)
							version := c.App.Metadata["version"].(*util.Version)

							action := &actions.UpdateInterop{
								Git:     git,
								GitHub:  github,
								Version: version,
							}

							return Execute(action, c)
						},
					},
					{
						Name:  "update-ipfs-docs",
						Usage: "Update the release in ipfs-docs",
						Action: func(c *cli.Context) error {
							github := c.App.Metadata["github"].(*github.Client)
							version := c.App.Metadata["version"].(*util.Version)

							action := &actions.UpdateIPFSDocs{
								GitHub:  github,
								Version: version,
							}

							return Execute(action, c)
						},
					},
					{
						Name:  "update-ipfs-blog",
						Usage: "Update the release in ipfs-blog",
						Flags: []cli.Flag{
							&cli.TimestampFlag{
								Name:    "date",
								Aliases: []string{"d"},
								Usage:   "Date of the release",
								Layout:  "2006-01-02",
							},
						},
						Action: func(c *cli.Context) error {
							git := c.App.Metadata["git"].(*git.Client)
							github := c.App.Metadata["github"].(*github.Client)
							version := c.App.Metadata["version"].(*util.Version)

							action := &actions.UpdateIPFSBlog{
								Git:     git,
								GitHub:  github,
								Version: version,
								Date:    c.Timestamp("date"),
							}

							return Execute(action, c)
						},
					},
					{
						Name:  "merge-branch",
						Usage: "Merge the release branch into master",
						Action: func(c *cli.Context) error {
							github := c.App.Metadata["github"].(*github.Client)
							version := c.App.Metadata["version"].(*util.Version)

							action := &actions.MergeBranch{
								GitHub:  github,
								Version: version,
							}

							return Execute(action, c)
						},
					},
				},
			},
		},
	}

	formatter := &log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	}
	log.SetFormatter(formatter)
	log.SetLevel(log.DebugLevel)

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

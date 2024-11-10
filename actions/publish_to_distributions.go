package actions

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/git"
	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
	log "github.com/sirupsen/logrus"
)

type PublishToDistributions struct {
	Git     *git.Client
	GitHub  *github.Client
	Version *util.Version
}

func (ctx PublishToDistributions) Check() error {
	log.Info("I'm going to check if the PR that publishes the release to distributions exists and if it's merged already.")

	err := CheckPR(ctx.GitHub, repos.Distributions.Owner, repos.Distributions.Repo, repos.Distributions.KuboBranch(ctx.Version), true)
	if err != nil {
		return err
	}

	return CheckBranch(ctx.GitHub, repos.Distributions.Owner, repos.Distributions.Repo, repos.Distributions.DefaultBranch)
}

func (ctx PublishToDistributions) Run() error {
	log.Info("I'm going to create a PR that publishes the release to distributions and ask you to merge it for me.")

	branch := repos.Distributions.KuboBranch(ctx.Version)
	title := fmt.Sprintf("Publish Kubo: %s", ctx.Version)
	body := fmt.Sprintf("This PR initiates publishing of Kubo %s", ctx.Version)

	b, err := ctx.GitHub.GetOrCreateBranch(repos.Distributions.Owner, repos.Distributions.Repo, branch, repos.Distributions.DefaultBranch)
	if err != nil {
		return err
	}

	err = ctx.Git.RunAndPush(repos.Distributions.Owner, repos.Distributions.Repo, branch, b.GetCommit().GetSHA(), "chore: add Kubo release", util.Command{Name: "./dist.sh", Args: []string{"add-version", "kubo", ctx.Version.Version}})
	if err != nil {
		return err
	}

	pr, err := ctx.GitHub.GetOrCreatePR(repos.Distributions.Owner, repos.Distributions.Repo, branch, repos.Distributions.DefaultBranch, title, body, false)
	if err != nil {
		return err
	}
	if !util.ConfirmPR(pr) {
		return fmt.Errorf("🚨 %s not merged", pr.GetHTMLURL())
	}
	return nil
}

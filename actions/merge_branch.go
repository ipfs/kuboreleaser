package actions

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
	log "github.com/sirupsen/logrus"
)

type MergeBranch struct {
	GitHub  *github.Client
	Version *util.Version
}

func (ctx MergeBranch) Check() error {
	log.Info("I'm going to check if the PR that merges the release branch to master exists and if it's merged already.")

	return CheckPR(ctx.GitHub, repos.Kubo.Owner, repos.Kubo.Repo, repos.Kubo.ReleaseMergeBranch(ctx.Version), true)
}

func (ctx MergeBranch) Run() error {
	log.Info("I'm going to create a PR that merges the release branch to master and ask you to merge it for me.")

	branch := repos.Kubo.ReleaseMergeBranch(ctx.Version)
	title := fmt.Sprintf("Merge Release: %s [skip changelog]", ctx.Version)
	body := fmt.Sprintf("This PR merges the release branch %s to %s", ctx.Version, repos.Kubo.DefaultBranch)

	_, err := ctx.GitHub.GetOrCreateBranch(repos.Kubo.Owner, repos.Kubo.Repo, branch, repos.Kubo.ReleaseBranch)
	if err != nil {
		return err
	}

	pr, err := ctx.GitHub.GetOrCreatePR(repos.Kubo.Owner, repos.Kubo.Repo, branch, repos.Kubo.DefaultBranch, title, body, false)
	if err != nil {
		return err
	}
	if !util.ConfirmPR(pr) {
		return fmt.Errorf("🚨 %s not merged", pr.GetHTMLURL())
	}

	return nil
}

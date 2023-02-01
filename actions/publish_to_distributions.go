package actions

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/git"
	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
)

type PublishToDistributions struct {
	Git     *git.Client
	GitHub  *github.Client
	Version *util.Version
}

func (ctx PublishToDistributions) Check() error {
	err := CheckPR(ctx.GitHub, repos.Distributions.Owner, repos.Distributions.Repo, repos.Distributions.KuboBranch(ctx.Version), true)
	if err != nil {
		return err
	}

	return CheckBranch(ctx.GitHub, repos.Distributions.Owner, repos.Distributions.Repo, repos.Distributions.DefaultBranch)
}

func (ctx PublishToDistributions) Run() error {
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
		return fmt.Errorf("pr not merged")
	}
	return nil
}

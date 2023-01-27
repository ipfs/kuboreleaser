package actions

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
)

type MergeBranch struct {
	GitHub  *github.Client
	Version *util.Version
}

func (ctx MergeBranch) Check() error {
	return CheckPR(ctx.GitHub, repos.Kubo.Owner, repos.Kubo.Repo, repos.Kubo.ReleaseMergeBranch(ctx.Version), true)
}

func (ctx MergeBranch) Run() error {
	branch := repos.Kubo.ReleaseMergeBranch(ctx.Version)
	title := fmt.Sprintf("Merge Release: %s", ctx.Version)
	body := fmt.Sprintf("This PR merges the release branch %s to %s", ctx.Version, repos.Kubo.DefaultBranch)

	_, err := ctx.GitHub.GetOrCreateBranch(repos.Kubo.Owner, repos.Kubo.Repo, branch, repos.Kubo.ReleaseBranch)
	if err != nil {
		return err
	}

	_, err = ctx.GitHub.GetOrCreatePR(repos.Kubo.Owner, repos.Kubo.Repo, branch, repos.Kubo.DefaultBranch, title, body, false)
	return err
}

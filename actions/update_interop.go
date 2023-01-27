package actions

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/git"
	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
)

type UpdateInterop struct {
	Git     *git.Client
	GitHub  *github.Client
	Version *util.Version
}

func (ctx UpdateInterop) Check() error {
	return CheckPR(ctx.GitHub, repos.Interop.Owner, repos.Interop.Repo, repos.Interop.KuboBranch(ctx.Version), !ctx.Version.IsPrerelease())
}

func (ctx UpdateInterop) Run() error {
	branch := repos.Interop.KuboBranch(ctx.Version)
	title := fmt.Sprintf("Update Kubo: %s", ctx.Version)
	body := fmt.Sprintf("This PR updates Kubo to %s", ctx.Version)
	command := git.Command{Name: "npm", Args: []string{"install", fmt.Sprintf("go-ipfs@%s", ctx.Version), "--save-dev"}}

	b, err := ctx.GitHub.GetOrCreateBranch(repos.Interop.Owner, repos.Interop.Repo, branch, repos.Interop.DefaultBranch)
	if err != nil {
		return err
	}

	err = ctx.Git.RunAndPush(repos.Interop.Owner, repos.Interop.Repo, branch, b.GetCommit().GetSHA(), "chore: update Kubo", command)
	if err != nil {
		return err
	}

	_, err = ctx.GitHub.GetOrCreatePR(repos.Interop.Owner, repos.Interop.Repo, branch, repos.Interop.DefaultBranch, title, body, false)
	return err
}

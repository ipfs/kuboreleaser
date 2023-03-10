package actions

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/git"
	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
)

type UpdateIPFSDesktop struct {
	Git     *git.Client
	GitHub  *github.Client
	Version *util.Version
}

func (ctx UpdateIPFSDesktop) Check() error {
	return CheckPR(ctx.GitHub, repos.IPFSDesktop.Owner, repos.IPFSDesktop.Repo, repos.IPFSDesktop.KuboBranch(ctx.Version), !ctx.Version.IsPrerelease())
}

func (ctx UpdateIPFSDesktop) Run() error {
	branch := repos.IPFSDesktop.KuboBranch(ctx.Version)
	title := fmt.Sprintf("Update Kubo: %s", ctx.Version)
	body := fmt.Sprintf("This PR updates Kubo to %s", ctx.Version)
	command := util.Command{Name: "npm", Args: []string{"install", fmt.Sprintf("go-ipfs@%s", ctx.Version), "--save"}}

	b, err := ctx.GitHub.GetOrCreateBranch(repos.IPFSDesktop.Owner, repos.IPFSDesktop.Repo, branch, repos.IPFSDesktop.DefaultBranch)
	if err != nil {
		return err
	}

	err = ctx.Git.RunAndPush(repos.IPFSDesktop.Owner, repos.IPFSDesktop.Repo, branch, b.GetCommit().GetSHA(), "chore: update Kubo", command)
	if err != nil {
		return err
	}

	pr, err := ctx.GitHub.GetOrCreatePR(repos.IPFSDesktop.Owner, repos.IPFSDesktop.Repo, branch, repos.IPFSDesktop.DefaultBranch, title, body, ctx.Version.IsPrerelease())
	if err != nil {
		return err
	}
	if !ctx.Version.IsPrerelease() {
		if !util.ConfirmPR(pr) {
			return fmt.Errorf("pr not merged")
		}
	}
	return nil
}

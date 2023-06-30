package actions

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/git"
	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
	log "github.com/sirupsen/logrus"
)

type UpdateInterop struct {
	Git     *git.Client
	GitHub  *github.Client
	Version *util.Version
}

func (ctx UpdateInterop) Check() error {
	log.Info("I'm going to check if the PR that updates Kubo in interop has been merged already.")

	return CheckPR(ctx.GitHub, repos.Interop.Owner, repos.Interop.Repo, repos.Interop.KuboBranch(ctx.Version), !ctx.Version.IsPrerelease())
}

func (ctx UpdateInterop) Run() error {
	log.Info("I'm going to create a PR that updates Kubo in interop.")

	branch := repos.Interop.KuboBranch(ctx.Version)
	title := fmt.Sprintf("Update Kubo: %s", ctx.Version)
	body := fmt.Sprintf("This PR updates Kubo to %s", ctx.Version)
	command := util.Command{Name: "npm", Args: []string{"install", fmt.Sprintf("go-ipfs@%s", ctx.Version), "--save-dev", "--save-exact"}}

	b, err := ctx.GitHub.GetOrCreateBranch(repos.Interop.Owner, repos.Interop.Repo, branch, repos.Interop.DefaultBranch)
	if err != nil {
		return err
	}

	err = ctx.Git.RunAndPush(repos.Interop.Owner, repos.Interop.Repo, branch, b.GetCommit().GetSHA(), "chore: update Kubo", command)
	if err != nil {
		return err
	}

	pr, err := ctx.GitHub.GetOrCreatePR(repos.Interop.Owner, repos.Interop.Repo, branch, repos.Interop.DefaultBranch, title, body, false)
	if err != nil {
		return err
	}
	if !util.ConfirmPR(pr) {
		return fmt.Errorf("%s not merged", pr.GetHTMLURL())
	}
	return nil
}

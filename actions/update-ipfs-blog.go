package actions

import (
	"fmt"
	"time"

	"github.com/ipfs/kuboreleaser/git"
	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
)

type UpdateIPFSBlog struct {
	Git     *git.Client
	GitHub  *github.Client
	Version *util.Version
	Date    *time.Time
}

func (ctx UpdateIPFSBlog) Check() error {
	return CheckPR(ctx.GitHub, repos.IPFSBlog.Owner, repos.IPFSBlog.Repo, repos.IPFSBlog.KuboBranch(ctx.Version), !ctx.Version.IsPrerelease())
}

func (ctx UpdateIPFSBlog) Run() error {
	branch := repos.IPFSBlog.KuboBranch(ctx.Version)
	title := fmt.Sprintf("Update Kubo: %s", ctx.Version)
	body := fmt.Sprintf("This PR updates Kubo to %s", ctx.Version)
	command := util.Command{Name: "yq", Args: []string{
		"ea",
		"-i",
		"-I", "0",
		"-N",
		fmt.Sprintf(`[.] | ["---", {"data": [{
			"title": "Just released: Kubo %s!",
			"date": "%s",
			"publish_date": null,
			"path": "https://github.com/ipfs/kubo/releases/tag/%s",
			"tags": [
				"go-ipfs",
				"kubo"
			]
		}]} *+ .[0], "---"] | .[]`, ctx.Version.String()[1:], ctx.Date.Format("2006-01-02"), ctx.Version.String()),
		"src/_blog/releasenotes.md",
	}}
	b, err := ctx.GitHub.GetOrCreateBranch(repos.IPFSBlog.Owner, repos.IPFSBlog.Repo, branch, repos.IPFSBlog.DefaultBranch)
	if err != nil {
		return err
	}

	err = ctx.Git.RunAndPush(repos.IPFSBlog.Owner, repos.IPFSBlog.Repo, branch, b.GetCommit().GetSHA(), "chore: add Kubo release note", command)
	if err != nil {
		return err
	}

	pr, err := ctx.GitHub.GetOrCreatePR(repos.IPFSBlog.Owner, repos.IPFSBlog.Repo, branch, repos.IPFSBlog.DefaultBranch, title, body, false)
	if err != nil {
		return err
	}
	if !util.ConfirmPR(pr) {
		return fmt.Errorf("pr not merged")
	}
	return nil
}

package actions

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/git"
	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
)

type Tag struct {
	Git     *git.Client
	GitHub  *github.Client
	Version *util.Version
}

func (ctx Tag) getBranch() string {
	if ctx.Version.IsPrerelease() {
		return repos.Kubo.VersionReleaseBranch(ctx.Version)
	} else {
		return repos.Kubo.ReleaseBranch
	}
}

func (ctx Tag) Check() error {
	tag, err := ctx.GitHub.GetTag(repos.Kubo.Owner, repos.Kubo.Repo, ctx.Version.String())
	if err != nil {
		return err
	}
	if tag == nil {
		return fmt.Errorf("tag %s does not exist (%w)", ctx.Version.String(), ErrIncomplete)
	}
	return nil
}

func (ctx Tag) Run() error {
	branch, err := ctx.GitHub.GetBranch(repos.Kubo.Owner, repos.Kubo.Repo, ctx.getBranch())
	if err != nil {
		return err
	}
	if branch == nil {
		return fmt.Errorf("branch %s does not exist", ctx.getBranch())
	}

	sha := branch.GetCommit().GetSHA()

	return ctx.Git.WithClone(repos.Kubo.Owner, repos.Kubo.Repo, branch.GetName(), sha, func(c *git.Clone) error {
		ref, err := c.Tag(sha, ctx.Version.String(), fmt.Sprintf("Release %s", ctx.Version))
		if err != nil {
			return err
		}

		prompt := fmt.Sprintf(`Tag created:
%v

Signature: %s

The tag will now be pushed to the remote repository.`, ref, ref.PGPSignature)
		if !util.Confirm(prompt) {
			return fmt.Errorf("tag creation aborted")
		}

		return c.PushTag(ctx.Version.String())
	})
}

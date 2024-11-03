package actions

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/git"
	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
	log "github.com/sirupsen/logrus"
)

type Tag struct {
	Git     *git.Client
	GitHub  *github.Client
	Version *util.Version
}

func (ctx Tag) getBranch() string {
	// NOTE: for patch releases (and prereleases), we should use the the version release branch because the release branch might be ahead already
	if ctx.Version.IsPrerelease() || ctx.Version.IsPatch() {
		return repos.Kubo.VersionReleaseBranch(ctx.Version)
	} else {
		return repos.Kubo.ReleaseBranch
	}
}

func (ctx Tag) Check() error {
	log.Info("I'm going to check if the signed tag for the release already exists.")

	tag, err := ctx.GitHub.GetTag(repos.Kubo.Owner, repos.Kubo.Repo, ctx.Version.String())
	if err != nil {
		return err
	}
	if tag == nil {
		return fmt.Errorf("⚠️ https://github.com/%s/%s/tags/%s does not exist (%w)", repos.Kubo.Owner, repos.Kubo.Repo, ctx.Version.String(), ErrIncomplete)
	}
	return nil
}

func (ctx Tag) Run() error {
	log.Info("I'm going to create a signed tag for the release.")

	branch, err := ctx.GitHub.GetBranch(repos.Kubo.Owner, repos.Kubo.Repo, ctx.getBranch())
	if err != nil {
		return err
	}
	if branch == nil {
		return fmt.Errorf("🚨 https://github.com/%s/%s/blob/%s does not exist", repos.Kubo.Owner, repos.Kubo.Repo, ctx.getBranch())
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

Please approve if the tag is correct. When you do, the tag will be pushed to the remote repository.`, ref, ref.PGPSignature)
		if !util.Confirm(prompt) {
			return fmt.Errorf("🚨 creation of tag '%s' was not confirmed correctly", ctx.Version.String())
		}

		return c.PushTag(ctx.Version.String())
	})
}

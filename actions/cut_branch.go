package actions

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/git"
	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
)

type CutBranch struct {
	Git     *git.Client
	GitHub  *github.Client
	Version *util.Version
}

func (ctx CutBranch) Check() error {
	versionReleaseBranch := repos.Kubo.VersionReleaseBranch(ctx.Version)
	versionUpdateBranch := repos.Kubo.VersionUpdateBranch(ctx.Version)

	err := CheckPR(ctx.GitHub, repos.Kubo.Owner, repos.Kubo.Repo, versionUpdateBranch, true)
	if err != nil {
		return err
	}

	return CheckPR(ctx.GitHub, repos.Kubo.Owner, repos.Kubo.Repo, versionReleaseBranch, !ctx.Version.IsPrerelease())
}

func (ctx CutBranch) UpdateVersion(branch, source, currentVersionNumber, base, title, body string, draft bool) error {
	b, err := ctx.GitHub.GetOrCreateBranch(repos.Kubo.Owner, repos.Kubo.Repo, branch, source)
	if err != nil {
		return err
	}

	err = ctx.Git.RunAndPush(repos.Kubo.Owner, repos.Kubo.Repo, branch, b.GetCommit().GetSHA(), "chore: update version", git.Command{Name: "sed", Args: []string{"-i", fmt.Sprintf("s/const CurrentVersionNumber = \".*\"/const CurrentVersionNumber = \"%s\"/g", currentVersionNumber), "version.go"}})
	if err != nil {
		return err
	}

	_, err = ctx.GitHub.GetOrCreatePR(repos.Kubo.Owner, repos.Kubo.Repo, branch, base, title, body, draft)
	return err
}

func (ctx CutBranch) Run() error {
	dev, err := ctx.Version.Dev()
	if err != nil {
		return err
	}

	branch := repos.Kubo.VersionReleaseBranch(ctx.Version)
	source := repos.Kubo.DefaultBranch
	currentVersionNumber := ctx.Version.String()[1:]
	base := repos.Kubo.ReleaseBranch
	title := fmt.Sprintf("Release: %s", ctx.Version.MajorMinorPatch())
	body := fmt.Sprintf("This PR creates release %s", ctx.Version.MajorMinorPatch())
	draft := ctx.Version.IsPrerelease()

	err = ctx.UpdateVersion(branch, source, currentVersionNumber, base, title, body, draft)
	if err != nil {
		return err
	}

	branch = repos.Kubo.VersionUpdateBranch(ctx.Version)
	source = repos.Kubo.DefaultBranch
	currentVersionNumber = dev[1:]
	base = repos.Kubo.DefaultBranch
	title = fmt.Sprintf("Update Version: %s", ctx.Version.MajorMinor())
	body = fmt.Sprintf("This PR updates version as part of the %s release", ctx.Version.MajorMinor())
	draft = false

	return ctx.UpdateVersion(branch, source, currentVersionNumber, base, title, body, draft)
}

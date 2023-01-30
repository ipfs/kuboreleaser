package actions

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/git"
	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
)

type PrepareBranch struct {
	Git     *git.Client
	GitHub  *github.Client
	Version *util.Version
}

func (ctx PrepareBranch) Check() error {
	versionReleaseBranch := repos.Kubo.VersionReleaseBranch(ctx.Version)

	err := CheckPR(ctx.GitHub, repos.Kubo.Owner, repos.Kubo.Repo, versionReleaseBranch, true)
	if err != nil {
		return err
	}

	if !ctx.Version.IsPatch() {
		versionUpdateBranch := repos.Kubo.VersionUpdateBranch(ctx.Version)
		err := CheckPR(ctx.GitHub, repos.Kubo.Owner, repos.Kubo.Repo, versionUpdateBranch, !ctx.Version.IsPrerelease())
		if err != nil {
			return err
		}
	}

	return nil
}

func (ctx PrepareBranch) UpdateVersion(branch, source, currentVersionNumber, base, title, body string, draft bool) error {
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

func (ctx PrepareBranch) Run() error {
	dev, err := ctx.Version.Dev()
	if err != nil {
		return err
	}

	branch := repos.Kubo.VersionReleaseBranch(ctx.Version)
	var source string
	if ctx.Version.IsPatch() {
		source = repos.Kubo.ReleaseBranch
	} else {
		source = repos.Kubo.DefaultBranch
	}
	currentVersionNumber := ctx.Version.String()[1:]
	base := repos.Kubo.ReleaseBranch
	title := fmt.Sprintf("Release: %s", ctx.Version.MajorMinorPatch())
	body := fmt.Sprintf("This PR creates release %s", ctx.Version.MajorMinorPatch())
	draft := ctx.Version.IsPrerelease()

	err = ctx.UpdateVersion(branch, source, currentVersionNumber, base, title, body, draft)
	if err != nil {
		return err
	}

	if !ctx.Version.IsPatch() {
		branch = repos.Kubo.VersionUpdateBranch(ctx.Version)
		source = repos.Kubo.DefaultBranch
		currentVersionNumber = dev[1:]
		base = repos.Kubo.DefaultBranch
		title = fmt.Sprintf("Update Version: %s", ctx.Version.MajorMinor())
		body = fmt.Sprintf("This PR updates version as part of the %s release", ctx.Version.MajorMinor())
		draft = false

		err := ctx.UpdateVersion(branch, source, currentVersionNumber, base, title, body, draft)
		if err != nil {
			return err
		}
	}

	return nil
}

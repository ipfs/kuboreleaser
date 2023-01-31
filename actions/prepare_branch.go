package actions

import (
	"fmt"

	gh "github.com/google/go-github/v48/github"
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

	err := CheckPR(ctx.GitHub, repos.Kubo.Owner, repos.Kubo.Repo, versionReleaseBranch, !ctx.Version.IsPrerelease())
	if err != nil {
		return err
	}

	if !ctx.Version.IsPatch() {
		versionUpdateBranch := repos.Kubo.VersionUpdateBranch(ctx.Version)
		err := CheckPR(ctx.GitHub, repos.Kubo.Owner, repos.Kubo.Repo, versionUpdateBranch, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ctx PrepareBranch) UpdateVersion(branch, source, currentVersionNumber, base, title, body string, draft bool) (*gh.PullRequest, error) {
	b, err := ctx.GitHub.GetOrCreateBranch(repos.Kubo.Owner, repos.Kubo.Repo, branch, source)
	if err != nil {
		return nil, err
	}

	err = ctx.Git.RunAndPush(repos.Kubo.Owner, repos.Kubo.Repo, branch, b.GetCommit().GetSHA(), "chore: update version", git.Command{Name: "sed", Args: []string{"-i", fmt.Sprintf("s/const CurrentVersionNumber = \".*\"/const CurrentVersionNumber = \"%s\"/g", currentVersionNumber), "version.go"}})
	if err != nil {
		return nil, err
	}

	return ctx.GitHub.GetOrCreatePR(repos.Kubo.Owner, repos.Kubo.Repo, branch, base, title, body, draft)
}

func (ctx PrepareBranch) Run() error {
	dev := fmt.Sprintf("%s.0-dev", ctx.Version.NextMajorMinor())

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

	pr, err := ctx.UpdateVersion(branch, source, currentVersionNumber, base, title, body, draft)
	if err != nil {
		return err
	}

	prompt := fmt.Sprintf(`If needed, check out the %s branch of %s/%s repository and cherry-pick commits from %s using the following command:

git cherry-pick -x <commit>

Please approve after all the required commits are cherry-picked.`, branch, repos.Kubo.Owner, repos.Kubo.Repo, repos.Kubo.DefaultBranch)
	if !util.Confirm(prompt) {
		return fmt.Errorf("cherry-picking commits from %s to %s was not confirmed", repos.Kubo.DefaultBranch, branch)
	}

	if !ctx.Version.IsPrerelease() {
		prompt := fmt.Sprintf(`Check out the %s branch of %s/%s repository and run the following command:

./bin/mkreleaselog 2>/dev/null

Now copy the stdout and update the Changelog and Contributors sections of the changelog file (https://github.com/%s/%s/blob/%s/docs/changelogs/%s.md).

Please approve after the changelog is updated.`, branch, repos.Kubo.Owner, repos.Kubo.Repo, repos.Kubo.Owner, repos.Kubo.Repo, branch, ctx.Version.MajorMinor())
		if !util.Confirm(prompt) {
			return fmt.Errorf("updating the changelog was not confirmed")
		}

		prompt = fmt.Sprintf(`Go to %s, ensure the CI checks pass, and merge the PR.

Please approve once the PR is merged.`, pr.GetHTMLURL())
		if !util.Confirm(prompt) {
			return fmt.Errorf("pr not merged")
		}
	}

	if !ctx.Version.IsPatch() {
		branch = repos.Kubo.VersionUpdateBranch(ctx.Version)
		source = repos.Kubo.DefaultBranch
		currentVersionNumber = dev[1:]
		base = repos.Kubo.DefaultBranch
		title = fmt.Sprintf("Update Version: %s", ctx.Version.MajorMinor())
		body = fmt.Sprintf("This PR updates version as part of the %s release", ctx.Version.MajorMinor())
		draft = false

		pr, err := ctx.UpdateVersion(branch, source, currentVersionNumber, base, title, body, draft)
		if err != nil {
			return err
		}
		prompt = fmt.Sprintf(`Go to %s, ensure the CI checks pass, and merge the PR.

Please approve once the PR is merged.`, pr.GetHTMLURL())
		if !util.Confirm(prompt) {
			return fmt.Errorf("pr not merged")
		}
	}

	return nil
}

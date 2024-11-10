package actions

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/ipfs/kuboreleaser/git"
	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
	log "github.com/sirupsen/logrus"
)

type PrepareNext struct {
	Git     *git.Client
	GitHub  *github.Client
	Version *util.Version
}

func (ctx PrepareNext) getNextVersion() *util.Version {
	version, _ := util.NewVersion(ctx.Version.NextMajorMinor())
	return version
}

func (ctx PrepareNext) Check() error {
	log.Info("I'm going to check if the PR that creates the next changelog exists and if it's merged already.")
	log.Info("I'm also going to check if the next release issue exists already.")

	next := ctx.getNextVersion()
	branch := repos.Kubo.ChangelogBranch(next)
	title := repos.Kubo.ReleaseIssueTitle(next)

	issue, err := ctx.GitHub.GetIssue(repos.Kubo.Owner, repos.Kubo.Repo, title)
	if err != nil {
		return err
	}
	if issue == nil {
		return fmt.Errorf("⚠️ issue '%s' not found in https://github.com/%s/%s/issues (%w)", title, repos.Kubo.Owner, repos.Kubo.Repo, ErrIncomplete)
	}

	err = CheckPR(ctx.GitHub, repos.Kubo.Owner, repos.Kubo.Repo, branch, true)
	if err != nil {
		return err
	}

	return nil
}

func (ctx PrepareNext) Run() error {
	log.Info("I'm going to create a PR that creates the next changelog and ask you to merge it for me.")
	log.Info("I'm also going to create the next release issue.")

	file, err := ctx.GitHub.GetFile(repos.Kubo.Owner, repos.Kubo.Repo, "docs/RELEASE_ISSUE_TEMPLATE.md", repos.Kubo.DefaultBranch)
	if err != nil {
		return err
	}
	if file == nil {
		return fmt.Errorf("🚨 https://github.com/%s/%s/tree/%s/docs/RELEASE_ISSUE_TEMPLATE.md not found", repos.Kubo.Owner, repos.Kubo.Repo, repos.Kubo.DefaultBranch)
	}

	content, err := base64.StdEncoding.DecodeString(*file.Content)
	if err != nil {
		return err
	}

	next := ctx.getNextVersion()
	branch := repos.Kubo.ChangelogBranch(next)
	issueTitle := repos.Kubo.ReleaseIssueTitle(next)
	issueBody := string(content)
	issueBody = strings.ReplaceAll(issueBody, "vX.Y.Z", next.MajorMinorPatch())
	issueBody = strings.ReplaceAll(issueBody, "vX.Y", next.MajorMinor())
	prTitle := fmt.Sprintf("Create Changelog: %s", next.MajorMinor())
	prBody := fmt.Sprintf("This PR creates changelog: %s", next.MajorMinor())

	_, err = ctx.GitHub.GetOrCreateIssue(repos.Kubo.Owner, repos.Kubo.Repo, issueTitle, issueBody)
	if err != nil {
		return err
	}

	b, err := ctx.GitHub.GetOrCreateBranch(repos.Kubo.Owner, repos.Kubo.Repo, branch, repos.Kubo.DefaultBranch)
	if err != nil {
		return err
	}

	createChangelog := util.Command{
		Name: "bash",
		Args: []string{"-c", fmt.Sprintf(`
			if [ ! -f docs/changelogs/%s.md ]; then
				cat > docs/changelogs/%s.md <<- EOF
					# Kubo changelog %s

					- [%s.0](#%s0)

					## %s.0

					- [Overview](#overview)
					- [🔦 Highlights](#-highlights)
					- [📝 Changelog](#-changelog)
					- [👨‍👩‍👧‍👦 Contributors](#-contributors)

					### Overview

					### 🔦 Highlights

					### 📝 Changelog

					### 👨‍👩‍👧‍👦 Contributors
				EOF
			fi
		`, next.MajorMinor(), next.MajorMinor(), next.MajorMinor(), next.MajorMinor(), fmt.Sprintf("%s%s", next.Major(), next.Minor()), next.MajorMinor())},
	}
	linkChangelog := util.Command{
		Name: "bash",
		Args: []string{"-c", fmt.Sprintf(`
			if ! grep -q %s CHANGELOG.md; then
				sed -i '3i - [%s](docs/changelogs/%s.md)' CHANGELOG.md
			fi
		`, next.MajorMinor(), next.MajorMinor(), next.MajorMinor())},
	}
	err = ctx.Git.RunAndPush(repos.Kubo.Owner, repos.Kubo.Repo, branch, b.GetCommit().GetSHA(), "chore: create next changelog", createChangelog, linkChangelog)
	if err != nil {
		return err
	}

	pr, err := ctx.GitHub.GetOrCreatePR(repos.Kubo.Owner, repos.Kubo.Repo, branch, repos.Kubo.DefaultBranch, prTitle, prBody, false)
	if err != nil {
		return err
	}

	if !util.ConfirmPR(pr) {
		return fmt.Errorf("🚨 %s not merged", pr.GetHTMLURL())
	}

	return nil
}

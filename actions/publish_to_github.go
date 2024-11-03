package actions

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
	log "github.com/sirupsen/logrus"
)

type PublishToGitHub struct {
	GitHub  *github.Client
	Version *util.Version
}

func (ctx PublishToGitHub) Check() error {
	log.Info("I'm going to check if the release has been created in GitHub and if the workflow that syncs the release assets has run already.")

	release, err := ctx.GitHub.GetRelease(repos.Kubo.Owner, repos.Kubo.Repo, ctx.Version.String())
	if err != nil {
		return err
	}
	if release == nil {
		return fmt.Errorf("release '%s' not found in https://github.com/%s/%s/releases (%w)", ctx.Version.String(), repos.Kubo.Owner, repos.Kubo.Repo, ErrIncomplete)
	}

	return CheckWorkflowRun(ctx.GitHub, repos.Kubo.Owner, repos.Kubo.Repo, repos.Kubo.DefaultBranch, repos.Kubo.SyncReleaseAssetsWorkflowName, repos.Kubo.SyncReleaseAssetsWorkflowJobName, ctx.Version.String())
}

func (ctx PublishToGitHub) Run() error {
	log.Info("I'm going to create a release in GitHub and a workflow run that syncs the release assets.")

	var body string
	if ctx.Version.IsPrerelease() {
		body = fmt.Sprintf("Changelog: [docs/changelogs/%s.md](https://github.com/ipfs/kubo/blob/release-%s/docs/changelogs/%s.md)", ctx.Version.MajorMinor(), ctx.Version.MajorMinorPatch(), ctx.Version.MajorMinor())
	} else {
		file, err := ctx.GitHub.GetFile(repos.Kubo.Owner, repos.Kubo.Repo, fmt.Sprintf("docs/changelogs/%s.md", ctx.Version.MajorMinor()), repos.Kubo.ReleaseBranch)
		if err != nil {
			return err
		}
		if file == nil {
			return fmt.Errorf("https://github.com/%s/%s/blob/%s/docs/changelogs/%s.md not found", repos.Kubo.Owner, repos.Kubo.Repo, repos.Kubo.ReleaseBranch, ctx.Version.MajorMinor())
		}

		content, err := base64.StdEncoding.DecodeString(*file.Content)
		if err != nil {
			return err
		}

		body = string(content)

		header := fmt.Sprintf("## %s\n", ctx.Version.MajorMinorPatch())

		index := strings.Index(body, header)
		if index != -1 {
			body = body[index+len(header):]
		} else {
			body = "<!-- Please fill out the release description manually -->"
		}

		index = strings.Index(body, "## ")
		if index != -1 {
			body = body[:index]
		}
	}

	_, err := ctx.GitHub.GetOrCreateRelease(repos.Kubo.Owner, repos.Kubo.Repo, ctx.Version.String(), ctx.Version.String(), body, ctx.Version.IsPrerelease())
	if err != nil {
		return err
	}

	return ctx.GitHub.CreateWorkflowRun(repos.Kubo.Owner, repos.Kubo.Repo, repos.Kubo.SyncReleaseAssetsWorkflowName, repos.Kubo.DefaultBranch)
}

package actions

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
)

type PublishToGitHub struct {
	GitHub  *github.Client
	Version *util.Version
}

func (ctx PublishToGitHub) Check() error {
	release, err := ctx.GitHub.GetRelease(repos.Kubo.Owner, repos.Kubo.Repo, ctx.Version.String())
	if err != nil {
		return err
	}
	if release == nil {
		return fmt.Errorf("release not found (%w)", ErrIncomplete)
	}

	return CheckWorkflowRun(ctx.GitHub, repos.Kubo.Owner, repos.Kubo.Repo, repos.Kubo.SyncReleaseAssetsWorkflowName, repos.Kubo.SyncReleaseAssetsWorkflowJobName, ctx.Version.String())
}

func (ctx PublishToGitHub) Run() error {
	var body string
	if ctx.Version.IsPrerelease() {
		body = fmt.Sprintf("Changelog: [docs/changelogs/%s.md](https://github.com/ipfs/kubo/blob/release-%s/docs/changelogs/%s.md)", ctx.Version.MajorMinor(), ctx.Version.MajorMinor(), ctx.Version.MajorMinor())
	} else {
		file, err := ctx.GitHub.GetFile(repos.Kubo.Owner, repos.Kubo.Repo, fmt.Sprintf("docs/changelogs/%s.md", ctx.Version.MajorMinor()), repos.Kubo.ReleaseBranch)
		if err != nil {
			return err
		}

		content, err := base64.StdEncoding.DecodeString(*file.Content)
		if err != nil {
			return err
		}

		body = string(content)

		index := strings.Index(body, "- [Overview](#overview)\n")
		if index != -1 {
			index += len("- [Overview](#overview)\n")
			body = body[index:]
		}

		if ctx.Version.IsPatch() {
			patch, err := strconv.Atoi(ctx.Version.Patch())
			if err != nil {
				return err
			}
			index = strings.Index(body, fmt.Sprintf("## %s.%d\n", ctx.Version.MajorMinor(), patch-1))
			if index != -1 {
				body = body[:index]
			}
		}
	}

	_, err := ctx.GitHub.GetOrCreateRelease(repos.Kubo.Owner, repos.Kubo.Repo, ctx.Version.String(), ctx.Version.String(), body, ctx.Version.IsPrerelease())
	if err != nil {
		return err
	}

	return ctx.GitHub.CreateWorkflowRun(repos.Kubo.Owner, repos.Kubo.Repo, repos.Kubo.SyncReleaseAssetsWorkflowName, repos.Kubo.DefaultBranch)
}

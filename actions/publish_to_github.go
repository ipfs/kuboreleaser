package actions

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/util"
)

type PublishToGitHub struct {
	github     *github.Client
	owner      string
	repo       string
	head       string
	base       string
	name       string
	majorMinor string
	prerelease bool
	workflow   string
	version    string
}

func NewPublishToGitHub(github *github.Client, version *util.Version) (*PublishToGitHub, error) {
	return &PublishToGitHub{
		github:     github,
		owner:      "ipfs",
		repo:       "kubo",
		head:       "release",
		base:       "master",
		workflow:   "sync-release-assets.yml",
		version:    version.Version,
		name:       version.Version,
		majorMinor: version.MajorMinor(),
		prerelease: version.Prerelease() != "",
	}, nil
}

func (ctx PublishToGitHub) Check() error {
	release, err := ctx.github.GetRelease(ctx.owner, ctx.repo, ctx.version)
	if err != nil {
		return err
	}

	if release == nil {
		return &util.CheckError{Action: util.CheckErrorRetry, Err: fmt.Errorf("release %s not found", ctx.version)}
	}

	run, err := ctx.github.GetWorkflowRun(ctx.owner, ctx.repo, ctx.workflow, false)
	if err != nil {
		return err
	}

	if run == nil {
		return &util.CheckError{Action: util.CheckErrorFail, Err: fmt.Errorf("no workflow runs found")}
	}

	if run.GetStatus() != "completed" {
		return &util.CheckError{Action: util.CheckErrorWait, Err: fmt.Errorf("the latest run is not completed")}
	}

	logs, err := ctx.github.GetWorkflowRunLogs(ctx.owner, ctx.repo, run.GetID())
	if err != nil {
		return err
	}

	sync := logs.JobLogs["sync-github-and-dist-ipfs-tech"]
	if sync == nil {
		return &util.CheckError{Action: util.CheckErrorFail, Err: fmt.Errorf("the latest run does not have a sync-github-and-dist-ipfs-tech job")}
	}

	if !strings.Contains(sync.RawLogs, ctx.version) {
		return &util.CheckError{Action: util.CheckErrorRetry, Err: fmt.Errorf("the latest run does not have the version %s", ctx.version)}
	}

	if run.GetConclusion() != "success" {
		return &util.CheckError{Action: util.CheckErrorFail, Err: fmt.Errorf("the latest run did not succeed")}
	}

	release, err = ctx.github.GetRelease(ctx.owner, ctx.repo, ctx.version)
	if err != nil {
		return err
	}

	if release.Assets == nil || len(release.Assets) == 0 {
		return &util.CheckError{Action: util.CheckErrorFail, Err: fmt.Errorf("the release does not have any assets")}
	}

	return nil
}

func (ctx PublishToGitHub) Run() error {
	var body string
	if ctx.prerelease {
		body = fmt.Sprintf("Changelog: [docs/changelogs/%s.md](https://github.com/ipfs/kubo/blob/release-%s/docs/changelogs/%s.md)", ctx.majorMinor, ctx.majorMinor, ctx.majorMinor)
	} else {
		file, err := ctx.github.GetFile(ctx.owner, ctx.repo, fmt.Sprintf("docs/changelogs/%s.md", ctx.majorMinor), ctx.head)
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
	}

	_, err := ctx.github.GetOrCreateRelease(ctx.owner, ctx.repo, ctx.version, ctx.name, body, ctx.prerelease)
	if err != nil {
		return err
	}

	return ctx.github.CreateWorkflowRun(ctx.owner, ctx.repo, ctx.workflow, ctx.base)
}

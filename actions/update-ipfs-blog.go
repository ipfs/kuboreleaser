package actions

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/ipfs/kuboreleaser/git"
	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/util"
)

type UpdateIPFSBlog struct {
	git     *git.Client
	github  *github.Client
	owner   string
	repo    string
	base    string
	head    string
	title   string
	body    string
	version string
	file    string
	date    *time.Time
}

func NewUpdateIPFSBlog(git *git.Client, github *github.Client, version *util.Version, date *time.Time) (*UpdateIPFSBlog, error) {
	return &UpdateIPFSBlog{
		git:     git,
		github:  github,
		owner:   "ipfs",
		repo:    "ipfs-blog",
		base:    "main",
		head:    fmt.Sprintf("update-kubo-%s", version.MajorMinor()),
		title:   fmt.Sprintf("Add release note: Kubo %s", version.Version),
		body:    fmt.Sprintf("This PR adds a release note for the %s Kubo release.", version.Version),
		version: version.Version,
		file:    "src/_blog/releasenotes.md",
		date:    date,
	}, nil
}

func (ctx UpdateIPFSBlog) Check() error {
	file, err := ctx.github.GetFile(ctx.owner, ctx.repo, ctx.file, ctx.base)
	if err != nil {
		return err
	}

	content, err := base64.StdEncoding.DecodeString(*file.Content)
	if err != nil {
		return err
	}

	if !strings.Contains(string(content[:]), fmt.Sprintf("Kubo %s!", ctx.version[1:])) {
		file, err = ctx.github.GetFile(ctx.owner, ctx.repo, ctx.file, ctx.head)
		if err != nil {
			return err
		}

		if file == nil {
			return &util.CheckError{
				Action: util.CheckErrorRetry,
				Err:    fmt.Errorf("file not found"),
			}
		}

		content, err := base64.StdEncoding.DecodeString(*file.Content)
		if err != nil {
			return err
		}

		if !strings.Contains(string(content[:]), fmt.Sprintf("Kubo %s!", ctx.version[1:])) {
			return &util.CheckError{Action: util.CheckErrorRetry, Err: fmt.Errorf("file does not contain version")}
		}

		pr, err := ctx.github.GetPR(ctx.owner, ctx.repo, ctx.head)
		if err != nil {
			return err
		}

		if pr == nil {
			return &util.CheckError{
				Action: util.CheckErrorRetry,
				Err:    fmt.Errorf("PR not found"),
			}
		}

		runs, err := ctx.github.GetCheckRuns(ctx.owner, ctx.repo, ctx.head)
		if err != nil {
			return err
		}

		for _, run := range runs {
			if run.GetStatus() == "completed" && run.GetConclusion() != "success" {
				return &util.CheckError{
					Action: util.CheckErrorFail,
					Err:    fmt.Errorf("check %s is not successful", run.GetName()),
				}
			}
		}

		for _, run := range runs {
			if run.GetStatus() != "completed" {
				return &util.CheckError{
					Action: util.CheckErrorWait,
					Err:    fmt.Errorf("check %s has not completed yet", run.GetName()),
				}
			}
		}
	}
	return nil
}

func (ctx UpdateIPFSBlog) Run() error {
	head, err := ctx.github.GetOrCreateBranch(ctx.owner, ctx.repo, ctx.head, ctx.base)
	if err != nil {
		return err
	}

	versionFile, err := ctx.github.GetFile(ctx.owner, ctx.repo, ctx.file, ctx.head)
	if err != nil {
		return err
	}

	content, err := base64.StdEncoding.DecodeString(*versionFile.Content)
	if err != nil {
		return err
	}

	if !strings.Contains(string(content[:]), fmt.Sprintf("Kubo %s!", ctx.version[1:])) {
		cmd :=
			git.Command{Name: "bash", Args: []string{
				"-c",
				fmt.Sprintf(`echo "---
$(cat '%s' |
	head -n-2 |
	tail -n+2 |
	yq --yaml-output --indentless '.data |= [
		{
			"title": "Just released: Kubo %s!",
			"date": "%s",
			"publish_date": null,
			"path": "https://github.com/ipfs/kubo/releases/tag/%s",
			"tags": [
				"go-ipfs",
				"kubo"
			]
		}
	] + .')
---" > '%s'`, ctx.file, ctx.version[1:], ctx.date.Format("2006-01-02"), ctx.version, ctx.file)}}
		err = ctx.git.WithCloneExecCommitAndPush(ctx.owner, ctx.repo, ctx.head, head.GetCommit().GetSHA(), ctx.file, fmt.Sprintf("chore: update %s", ctx.file), cmd)
		if err != nil {
			return err
		}
	}

	_, err = ctx.github.GetOrCreatePR(ctx.owner, ctx.repo, ctx.head, ctx.base, ctx.title, ctx.body, false)
	if err != nil {
		return err
	}

	return nil
}

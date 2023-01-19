package actions

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ipfs/kuboreleaser/git"
	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/util"
)

type PublishToDistributions struct {
	git          *git.Client
	github       *github.Client
	owner        string
	repo         string
	base         string
	head         string
	title        string
	body         string
	version      string
	versionsFile string
	glob         string
	distFile     string
}

func NewPublishToDistributions(git *git.Client, github *github.Client, version *util.Version) (*PublishToDistributions, error) {
	return &PublishToDistributions{
		git:          git,
		github:       github,
		owner:        "ipfs",
		repo:         "distributions",
		base:         "master",
		head:         fmt.Sprintf("publish-kubo-%s", version.Version),
		title:        fmt.Sprintf("Publish Kubo: %s", version.Version),
		body:         fmt.Sprintf("This PR initiates publishing of Kubo %s", version.Version),
		version:      version.Version,
		versionsFile: "dists/kubo/versions",
		glob:         "dists/*/versions",
		distFile:     "dist.sh",
	}, nil
}

func (ctx PublishToDistributions) Check() error {
	versionsFile, err := ctx.github.GetFile(ctx.owner, ctx.repo, ctx.versionsFile, ctx.base)
	if err != nil {
		return err
	}

	if strings.Contains(*versionsFile.Content, ctx.version) {
		runs, err := ctx.github.GetCheckRuns(ctx.owner, ctx.repo, ctx.base)
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

		response, err := http.Get("https://dist.ipfs.tech/kubo/versions")
		if err != nil {
			return err
		}

		versions, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}

		if !strings.Contains(string(versions), ctx.version) {
			return &util.CheckError{Action: util.CheckErrorFail, Err: fmt.Errorf("version %s not found in dist.ipfs.tech/kubo/versions", ctx.version)}
		}
	} else {
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

		if !pr.GetMerged() {
			return &util.CheckError{
				Action: util.CheckErrorWait,
				Err:    fmt.Errorf("PR is not merged yet"),
			}
		}
	}

	return nil
}

func (ctx PublishToDistributions) Run() error {
	versionsFile, err := ctx.github.GetFile(ctx.owner, ctx.repo, ctx.versionsFile, ctx.base)
	if err != nil {
		return err
	}

	if !strings.Contains(*versionsFile.Content, ctx.version) {
		head, err := ctx.github.GetOrCreateBranch(ctx.owner, ctx.repo, ctx.head, ctx.base)
		if err != nil {
			return err
		}

		versionsFile, err := ctx.github.GetFile(ctx.owner, ctx.repo, ctx.versionsFile, ctx.head)
		if err != nil {
			return err
		}

		if !strings.Contains(*versionsFile.Content, ctx.version) {
			cmd := git.Command{Name: ctx.distFile, Args: []string{"add-version", "kubo", ctx.version}}
			err = ctx.git.CloneExecCommitAndPush(ctx.owner, ctx.repo, ctx.head, head.GetCommit().GetSHA(), ctx.glob, fmt.Sprintf("chore: update %s", ctx.versionsFile), cmd)
			if err != nil {
				return err
			}
		}

		_, err = ctx.github.GetOrCreatePR(ctx.owner, ctx.repo, ctx.head, ctx.base, ctx.title, ctx.body, false)
		if err != nil {
			return err
		}
	}

	return nil
}

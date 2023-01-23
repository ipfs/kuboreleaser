package actions

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/util"
)

type MergeBranch struct {
	github *github.Client
	owner  string
	repo   string
	base   string
	head   string
	title  string
	body   string
}

func NewMergeBranch(github *github.Client, version *util.Version) (*MergeBranch, error) {
	return &MergeBranch{
		github: github,
		owner:  "ipfs",
		repo:   "kubo",
		base:   "master",
		head:   fmt.Sprintf("merge-kubo-%s", version.MajorMinor()),
		title:  fmt.Sprintf("Merge Kubo: %s", version.MajorMinor()),
		body:   fmt.Sprintf("This PR merges Kubo %s to master", version.MajorMinor()),
	}, nil
}

func (ctx MergeBranch) Check() error {
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

	return nil
}

func (ctx MergeBranch) Run() error {
	_, err := ctx.github.GetOrCreateBranch(ctx.owner, ctx.repo, ctx.head, "release")
	if err != nil {
		return err
	}

	_, err = ctx.github.GetOrCreatePR(ctx.owner, ctx.repo, ctx.head, ctx.base, ctx.title, ctx.body, false)
	if err != nil {
		return err
	}

	return nil
}

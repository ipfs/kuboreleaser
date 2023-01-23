package actions

import (
	"fmt"
	"strings"

	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/util"
)

type UpdateIPFSDocs struct {
	github   *github.Client
	owner    string
	repo     string
	base     string
	workflow string
	version  string
}

func NewUpdateIPFSDocs(github *github.Client, version *util.Version) (*UpdateIPFSDocs, error) {
	return &UpdateIPFSDocs{
		github:   github,
		owner:    "ipfs",
		repo:     "ipfs-docs",
		base:     "main",
		workflow: "update-on-new-ipfs-tag.yml",
		version:  version.Version,
	}, nil
}

func (ctx UpdateIPFSDocs) Check() error {
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

	update := logs.JobLogs["update"]
	if update == nil {
		return &util.CheckError{Action: util.CheckErrorFail, Err: fmt.Errorf("the latest run does not have an update job")}
	}

	if !strings.Contains(update.RawLogs, fmt.Sprintf(" %s\r\n", ctx.version)) {
		return &util.CheckError{Action: util.CheckErrorRetry, Err: fmt.Errorf("the latest run does not have the version %s", ctx.version)}
	}

	if run.GetConclusion() != "success" {
		return &util.CheckError{Action: util.CheckErrorFail, Err: fmt.Errorf("the latest run did not succeed")}
	}

	return nil
}

func (ctx UpdateIPFSDocs) Run() error {
	return ctx.github.CreateWorkflowRun(ctx.owner, ctx.repo, ctx.workflow, ctx.base)
}

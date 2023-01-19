package actions

import (
	"fmt"
	"strings"

	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/util"
)

type PublishToNPM struct {
	github   *github.Client
	owner    string
	repo     string
	base     string
	workflow string
	version  string
}

func NewPublishToNPM(github *github.Client, version *util.Version) (*PublishToNPM, error) {
	return &PublishToNPM{
		github:   github,
		owner:    "ipfs",
		repo:     "npm-go-ipfs",
		base:     "master",
		workflow: "main.yml",
		version:  version.Version,
	}, nil
}

func (ctx PublishToNPM) Check() error {
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

	publish := logs.JobLogs["publish"]
	if publish == nil {
		return &util.CheckError{Action: util.CheckErrorFail, Err: fmt.Errorf("the latest run does not have a publish job")}
	}

	if !strings.Contains(publish.RawLogs, ctx.version) {
		return &util.CheckError{Action: util.CheckErrorRetry, Err: fmt.Errorf("the latest run does not have the version %s", ctx.version)}
	}

	if run.GetConclusion() != "success" {
		return &util.CheckError{Action: util.CheckErrorFail, Err: fmt.Errorf("the latest run did not succeed")}
	}

	return nil
}

func (ctx PublishToNPM) Run() error {
	return ctx.github.CreateWorkflowRun(ctx.owner, ctx.repo, ctx.workflow, ctx.base)
}

package actions

import (
	"fmt"
	"strings"

	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/util"
)

type TestIPFSCompanion struct {
	github   *github.Client
	owner    string
	repo     string
	base     string
	workflow string
	version  string
}

func NewTestIPFSCompanion(github *github.Client, version *util.Version) (*TestIPFSCompanion, error) {
	return &TestIPFSCompanion{
		github:   github,
		owner:    "ipfs",
		repo:     "ipfs-companion",
		base:     "main",
		workflow: "e2e.yml",
		version:  version.Version,
	}, nil
}

func (ctx TestIPFSCompanion) Check() error {
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

	test := logs.JobLogs["test"]
	if test == nil {
		return &util.CheckError{Action: util.CheckErrorFail, Err: fmt.Errorf("the latest run does not have a test job")}
	}

	if !strings.Contains(test.RawLogs, ctx.version) {
		return &util.CheckError{Action: util.CheckErrorRetry, Err: fmt.Errorf("the latest run does not have the version %s", ctx.version)}
	}

	if run.GetConclusion() != "success" {
		return &util.CheckError{Action: util.CheckErrorFail, Err: fmt.Errorf("the latest run did not succeed")}
	}

	return nil
}

func (ctx TestIPFSCompanion) Run() error {
	return ctx.github.CreateWorkflowRun(ctx.owner, ctx.repo, ctx.workflow, ctx.base, github.WorkflowRunInput{Name: "kubo-version", Value: ctx.version})
}

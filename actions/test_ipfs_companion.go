package actions

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
	log "github.com/sirupsen/logrus"
)

type TestIPFSCompanion struct {
	GitHub  *github.Client
	Version *util.Version
}

func (ctx TestIPFSCompanion) Check() error {
	log.Info("I'm going to check if the workflow that tests IPFS Companion has run already.")

	return CheckWorkflowRun(ctx.GitHub, repos.IPFSCompanion.Owner, repos.IPFSCompanion.Repo, repos.IPFSCompanion.DefaultBranch, repos.IPFSCompanion.WorkflowName, repos.IPFSCompanion.WorkflowJobName, fmt.Sprintf(" %s\r\n", ctx.Version.String()))
}

func (ctx TestIPFSCompanion) Run() error {
	log.Info("I'm going to create a workflow run that tests IPFS Companion.")

	return ctx.GitHub.CreateWorkflowRun(repos.IPFSCompanion.Owner, repos.IPFSCompanion.Repo, repos.IPFSCompanion.WorkflowName, repos.IPFSCompanion.DefaultBranch, github.WorkflowRunInput{Name: "kubo-version", Value: ctx.Version.String()})
}

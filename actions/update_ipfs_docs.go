package actions

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
	log "github.com/sirupsen/logrus"
)

type UpdateIPFSDocs struct {
	GitHub  *github.Client
	Version *util.Version
}

func (ctx UpdateIPFSDocs) Check() error {
	log.Info("I'm going to check if the workflow that updates the IPFS docs has run already.")

	return CheckWorkflowRun(ctx.GitHub, repos.IPFSDocs.Owner, repos.IPFSDocs.Repo, repos.IPFSDocs.DefaultBranch, repos.IPFSDocs.WorkflowName, repos.IPFSDocs.WorkflowJobName, fmt.Sprintf(" %s\r\n", ctx.Version.String()))
}

func (ctx UpdateIPFSDocs) Run() error {
	log.Info("I'm going to create a workflow run that updates the IPFS docs.")

	return ctx.GitHub.CreateWorkflowRun(repos.IPFSDocs.Owner, repos.IPFSDocs.Repo, repos.IPFSDocs.WorkflowName, repos.IPFSDocs.DefaultBranch)
}

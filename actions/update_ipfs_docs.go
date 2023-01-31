package actions

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
)

type UpdateIPFSDocs struct {
	GitHub  *github.Client
	Version *util.Version
}

func (ctx UpdateIPFSDocs) Check() error {
	return CheckWorkflowRun(ctx.GitHub, repos.IPFSDocs.Owner, repos.IPFSDocs.Repo, repos.IPFSDocs.DefaultBranch, repos.IPFSDocs.WorkflowName, repos.IPFSDocs.WorkflowJobName, fmt.Sprintf(" %s\r\n", ctx.Version.String()))
}

func (ctx UpdateIPFSDocs) Run() error {
	return ctx.GitHub.CreateWorkflowRun(repos.IPFSDocs.Owner, repos.IPFSDocs.Repo, repos.IPFSDocs.WorkflowName, repos.IPFSDocs.DefaultBranch)
}

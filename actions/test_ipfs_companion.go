package actions

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
)

type TestIPFSCompanion struct {
	GitHub  *github.Client
	Version *util.Version
}

func (ctx TestIPFSCompanion) Check() error {
	return CheckWorkflowRun(ctx.GitHub, repos.IPFSCompanion.Owner, repos.IPFSCompanion.Repo, repos.IPFSCompanion.WorkflowName, repos.IPFSCompanion.WorkflowJobName, fmt.Sprintf(" %s\r\n", ctx.Version.String()))
}

func (ctx TestIPFSCompanion) Run() error {
	return ctx.GitHub.CreateWorkflowRun(repos.IPFSCompanion.Owner, repos.IPFSCompanion.Repo, repos.IPFSCompanion.WorkflowName, repos.IPFSCompanion.DefaultBranch)
}

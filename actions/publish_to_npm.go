package actions

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
)

type PublishToNPM struct {
	GitHub  *github.Client
	Version *util.Version
}

func (ctx PublishToNPM) Check() error {
	return CheckWorkflowRun(ctx.GitHub, repos.NPMGoIPFS.Owner, repos.NPMGoIPFS.Repo, repos.NPMGoIPFS.DefaultBranch, repos.NPMGoIPFS.WorkflowName, repos.NPMGoIPFS.WorkflowJobName, fmt.Sprintf(" %s\r\n", ctx.Version.String()[1:]))
}

func (ctx PublishToNPM) Run() error {
	return ctx.GitHub.CreateWorkflowRun(repos.NPMGoIPFS.Owner, repos.NPMGoIPFS.Repo, repos.NPMGoIPFS.WorkflowName, repos.NPMGoIPFS.DefaultBranch)
}

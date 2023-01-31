package actions

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
)

type PublishToDockerHub struct {
	GitHub  *github.Client
	Version *util.Version
}

func (ctx PublishToDockerHub) getBranch() string {
	if ctx.Version.IsPrerelease() {
		return repos.Kubo.VersionReleaseBranch(ctx.Version)
	}
	return repos.Kubo.ReleaseBranch
}

func (ctx PublishToDockerHub) Check() error {
	return CheckWorkflowRun(ctx.GitHub, repos.Kubo.Owner, repos.Kubo.Repo, ctx.getBranch(), repos.Kubo.DockerHubWorkflowName, repos.Kubo.DockerHubWorkflowJobName, fmt.Sprintf("ipfs/kubo:%s", ctx.Version))
}

func (ctx PublishToDockerHub) Run() error {
	return ctx.GitHub.CreateWorkflowRun(repos.Kubo.Owner, repos.Kubo.Repo, repos.Kubo.DockerHubWorkflowName, ctx.getBranch())
}

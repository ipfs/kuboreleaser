package actions

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
	log "github.com/sirupsen/logrus"
)

type PublishToDockerHub struct {
	GitHub  *github.Client
	Version *util.Version
}

func (ctx PublishToDockerHub) Check() error {
	log.Info("I'm going to check if the workflow that publishes the Docker image to Docker Hub has run already.")

	return CheckWorkflowRun(ctx.GitHub, repos.Kubo.Owner, repos.Kubo.Repo, ctx.Version.Version, repos.Kubo.DockerHubWorkflowName, repos.Kubo.DockerHubWorkflowJobName, fmt.Sprintf("ipfs/kubo:%s", ctx.Version))
}

func (ctx PublishToDockerHub) Run() error {
	log.Info("I'm going to create a workflow run that publishes the Docker image to Docker Hub.")

	return ctx.GitHub.CreateWorkflowRun(repos.Kubo.Owner, repos.Kubo.Repo, repos.Kubo.DockerHubWorkflowName, ctx.Version.Version)
}

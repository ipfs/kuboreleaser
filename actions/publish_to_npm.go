package actions

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
	log "github.com/sirupsen/logrus"
)

type PublishToNPM struct {
	GitHub  *github.Client
	Version *util.Version
}

func (ctx PublishToNPM) Check() error {
	log.Info("I'm going to check if the workflow that publishes the NPM package has run already.")

	return CheckWorkflowRun(ctx.GitHub, repos.NPMKubo.Owner, repos.NPMKubo.Repo, repos.NPMKubo.DefaultBranch, repos.NPMKubo.WorkflowName, repos.NPMKubo.WorkflowJobName, fmt.Sprintf(" %s\n", ctx.Version.String()[1:]))
}

func (ctx PublishToNPM) Run() error {
	log.Info("I'm going to create a workflow run that publishes the NPM package.")

	return ctx.GitHub.CreateWorkflowRun(repos.NPMKubo.Owner, repos.NPMKubo.Repo, repos.NPMKubo.WorkflowName, repos.NPMKubo.DefaultBranch)
}

package actions

import (
	"fmt"
	"time"

	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
)

type NotifyBifrost struct {
	GitHub  *github.Client
	Version *util.Version
	Date    *time.Time
}

func (ctx NotifyBifrost) getIssueTitle() string {
	return fmt.Sprintf("Rollout Kubo %s to a Cluster, Gateway and Bootstrapper bank", ctx.Version.MajorMinorPatch())
}

func (ctx NotifyBifrost) getIssueBody() string {
	return fmt.Sprintf("A new Kubo release process - %s - is starting on %s", ctx.Version.MajorMinorPatch(), ctx.Date.Format("2006-01-02"))
}

func (ctx NotifyBifrost) getIssueComment() string {
	return fmt.Sprintf("A new Kubo release - %s - was published on %s", ctx.Version, ctx.Date.Format("2006-01-02"))
}

func (ctx NotifyBifrost) isAdvanceNotice() bool {
	return ctx.Date.After(time.Now())
}

func (ctx NotifyBifrost) Check() error {
	issue, err := ctx.GitHub.GetIssue(repos.BifrostInfra.Owner, repos.BifrostInfra.Repo, ctx.getIssueTitle())
	if err != nil {
		return err
	}
	if issue == nil {
		return fmt.Errorf("issue %s not found (%w)", ctx.getIssueTitle(), ErrIncomplete)
	}

	if !ctx.isAdvanceNotice() {
		comment, err := ctx.GitHub.GetIssueComment(repos.BifrostInfra.Owner, repos.BifrostInfra.Repo, issue.GetNumber(), ctx.getIssueComment())
		if err != nil {
			return err
		}
		if comment == nil {
			return fmt.Errorf("comment %s not found (%w)", ctx.getIssueComment(), ErrIncomplete)
		}
	}

	return nil
}

func (ctx NotifyBifrost) Run() error {
	issue, err := ctx.GitHub.GetOrCreateIssue(repos.BifrostInfra.Owner, repos.BifrostInfra.Repo, ctx.getIssueTitle(), ctx.getIssueBody())
	if err != nil {
		return err
	}

	if !ctx.isAdvanceNotice() {
		_, err = ctx.GitHub.GetOrCreateIssueComment(repos.BifrostInfra.Owner, repos.BifrostInfra.Repo, issue.GetNumber(), ctx.getIssueComment())
		if err != nil {
			return err
		}
	}

	return nil
}

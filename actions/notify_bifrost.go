package actions

import (
	"fmt"
	"time"

	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/util"
)

type NotifyBifrost struct {
	github          *github.Client
	owner           string
	repo            string
	title           string
	body            string
	isAdvanceNotice bool
}

func NewNotifyBifrost(github *github.Client, version *util.Version, date *time.Time) (*NotifyBifrost, error) {
	title := fmt.Sprintf("Rollout Kubo %s to a Cluster, Gateway and Bootstrapper bank", version.MajorMinor())

	isAdvanceNotice := date.After(time.Now())

	var body string
	if isAdvanceNotice {
		body = fmt.Sprintf("A new Kubo release - %s - is going to be published on %s", version.Version, date.Format("2006-01-02"))
	} else {
		body = fmt.Sprintf("A new Kubo release - %s - was published on %s", version.Version, date.Format("2006-01-02"))
	}

	return &NotifyBifrost{
		github:          github,
		owner:           "protocol",
		repo:            "bifrost-infra",
		title:           title,
		body:            body,
		isAdvanceNotice: isAdvanceNotice,
	}, nil
}

func (ctx NotifyBifrost) Check() error {
	issue, err := ctx.github.GetIssue(ctx.owner, ctx.repo, ctx.title)
	if err != nil {
		return err
	}

	if issue == nil {
		return &util.CheckError{Action: util.CheckErrorRetry, Err: fmt.Errorf("issue %s not found", ctx.title)}
	}

	if !ctx.isAdvanceNotice {
		comment, err := ctx.github.GetIssueComment(ctx.owner, ctx.repo, issue.GetNumber(), ctx.body)
		if err != nil {
			return err
		}

		if comment == nil {
			return &util.CheckError{Action: util.CheckErrorRetry, Err: fmt.Errorf("comment %s not found", ctx.body)}
		}
	}

	return nil
}

func (ctx NotifyBifrost) Run() error {
	issue, err := ctx.github.GetOrCreateIssue(ctx.owner, ctx.repo, ctx.title, ctx.body)
	if err != nil {
		return err
	}

	if !ctx.isAdvanceNotice {
		_, err = ctx.github.GetOrCreateIssueComment(ctx.owner, ctx.repo, issue.GetNumber(), ctx.body)
		return err
	}

	return nil
}

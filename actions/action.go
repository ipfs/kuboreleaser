package actions

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/ipfs/kuboreleaser/github"
)

type IAction interface {
	Run() error
	Check() error
}

var (
	ErrInProgress = errors.New("the action is in progress")
	ErrIncomplete = errors.New("the action is not complete")
	ErrFailure    = errors.New("the action failed and requires manual intervention")
)

func CheckBranch(github *github.Client, owner, repo, branch string) error {
	runs, err := github.GetIncompleteCheckRuns(owner, repo, branch)
	if err != nil {
		return err
	}
	if len(runs) > 0 {
		return fmt.Errorf("⚠️ check %s on https://github.com/%s/%s/tree/%s is not completed yet (%w)", runs[0].GetName(), owner, repo, branch, ErrInProgress)
	}

	runs, err = github.GetUnsuccessfulCheckRuns(owner, repo, branch)
	if err != nil {
		return err
	}
	if len(runs) > 0 {
		return fmt.Errorf("⚠️ check %s on https://github.com/%s/%s/tree/%s is not successful (%w)", runs[0].GetName(), owner, repo, branch, ErrIncomplete)
	}

	return nil
}

func CheckPR(github *github.Client, owner, repo, head string, shouldBeMerged bool) error {
	pr, err := github.GetPR(owner, repo, head)
	if err != nil {
		return err
	}
	if pr == nil {
		return fmt.Errorf("⚠️ PR for https://github.com/%s/%s/tree/%s not found (%w)", owner, repo, head, ErrIncomplete)
	}

	if !pr.GetMerged() {
		if pr.GetState() == "closed" {
			return fmt.Errorf("⚠️ %s is closed (%w)", pr.GetHTMLURL(), ErrIncomplete)
		}

		err = CheckBranch(github, owner, repo, head)
		if err != nil {
			return err
		}

		if shouldBeMerged {
			return fmt.Errorf("⚠️ %s is not merged (%w)", pr.GetHTMLURL(), ErrInProgress)
		}
	}

	return nil
}

func CheckWorkflowRun(github *github.Client, owner, repo, branch, file, job, pattern string) error {
	run, err := github.GetWorkflowRun(owner, repo, branch, file, false)
	if err != nil {
		return err
	}
	if run == nil {
		return fmt.Errorf("⚠️ workflow run %s for https://github.com/%s/%s/tree/%s not found (%w)", file, owner, repo, branch, ErrIncomplete)
	}

	if run.GetStatus() != "completed" {
		return fmt.Errorf("⚠️ %s is not completed (%w)", run.GetHTMLURL(), ErrInProgress)
	}
	if run.GetConclusion() != "success" {
		return fmt.Errorf("⚠️ %s did not succeed (%w)", run.GetHTMLURL(), ErrFailure)
	}

	runLogs, err := github.GetWorkflowRunLogs(owner, repo, run.GetID())
	if err != nil {
		return err
	}

	jobLogs := runLogs.JobLogs[job]
	if jobLogs == nil {
		return fmt.Errorf("⚠️ %s does not have a %s job (%w)", run.GetHTMLURL(), job, ErrFailure)
	}

	matched, err := regexp.MatchString(pattern, jobLogs.RawLogs)
	if err != nil {
		return err
	}

	if !matched {
		return fmt.Errorf("⚠️ %s does not have the pattern %s (%w)", run.GetHTMLURL(), pattern, ErrIncomplete)
	}

	return nil
}

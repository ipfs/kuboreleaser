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
	ErrIncomplete = errors.New("invalid argument")
	ErrFailure    = errors.New("operation failed")
	ErrError      = errors.New("operation failed irrecoverably")
)

func CheckBranch(github *github.Client, owner, repo, branch string) error {
	runs, err := github.GetIncompleteCheckRuns(owner, repo, branch)
	if err != nil {
		return err
	}
	if len(runs) > 0 {
		return fmt.Errorf("check %s is not completed yet (%w)", runs[0].GetName(), ErrIncomplete)
	}

	runs, err = github.GetUnsuccessfulCheckRuns(owner, repo, branch)
	if err != nil {
		return err
	}
	if len(runs) > 0 {
		return fmt.Errorf("check %s is not successful (%w)", runs[0].GetName(), ErrFailure)
	}

	return nil
}

func CheckPR(github *github.Client, owner, repo, head string, shouldBeMerged bool) error {
	pr, err := github.GetPR(owner, repo, head)
	if err != nil {
		return err
	}
	if pr == nil {
		return fmt.Errorf("PR not found (%w)", ErrFailure)
	}

	err = CheckBranch(github, owner, repo, head)
	if err != nil {
		return err
	}

	if shouldBeMerged && !pr.GetMerged() {
		return fmt.Errorf("PR is not merged (%w)", ErrIncomplete)
	}

	return nil
}

func CheckWorkflowRun(github *github.Client, owner, repo, file, job, pattern string) error {
	run, err := github.GetWorkflowRun(owner, repo, file, false)
	if err != nil {
		return err
	}
	if run == nil {
		return fmt.Errorf("workflow run not found (%w)", ErrFailure)
	}
	if run.GetStatus() != "completed" {
		return fmt.Errorf("the latest run is not completed (%w)", ErrIncomplete)
	}
	if run.GetConclusion() != "success" {
		return fmt.Errorf("the latest run did not succeed (%w)", ErrError)
	}

	runLogs, err := github.GetWorkflowRunLogs(owner, repo, run.GetID())
	if err != nil {
		return err
	}

	jobLogs := runLogs.JobLogs[job]
	if jobLogs == nil {
		return fmt.Errorf("the latest run does not have a %s job (%w)", job, ErrError)
	}

	matched, err := regexp.MatchString(pattern, jobLogs.RawLogs)
	if err != nil {
		return err
	}

	if !matched {
		return fmt.Errorf("the latest run does not have the pattern %s (%w)", pattern, ErrFailure)
	}

	return nil
}

package actions

import (
	"fmt"
	"strings"

	"github.com/ipfs/kuboreleaser/git"
	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/util"
)

type CutBranch struct {
	git            *git.Client
	github         *github.Client
	owner          string
	repo           string
	defaultBase    string
	defaultHead    string
	defaultTitle   string
	defaultBody    string
	releaseBase    string
	releaseHead    string
	releaseTitle   string
	releaseBody    string
	defaultVersion string
	releaseVersion string
	versionFile    string
}

func NewCutBranch(git *git.Client, github *github.Client, version *util.Version) (*CutBranch, error) {
	devVersion, err := version.Dev()
	if err != nil {
		return nil, err
	}

	return &CutBranch{
		git:            git,
		github:         github,
		owner:          "ipfs",
		repo:           "kubo",
		defaultBase:    "master",
		releaseBase:    "release",
		defaultHead:    fmt.Sprintf("dev-version-update-%s", version.MajorMinor()),
		releaseHead:    fmt.Sprintf("release-%s", version.MajorMinor()),
		defaultTitle:   fmt.Sprintf("Dev Version Update: %s", version.MajorMinor()),
		releaseTitle:   fmt.Sprintf("Release: %s", version.MajorMinor()),
		defaultBody:    fmt.Sprintf("This PR updates version to %s", version.MajorMinor()),
		releaseBody:    fmt.Sprintf("This PR creates release %s", version.MajorMinor()),
		defaultVersion: devVersion[1:],      // check for v prefix
		releaseVersion: version.Version[1:], // check for v prefix
		versionFile:    "version.go",
	}, nil
}

func (ctx CutBranch) Check() error {
	file, err := ctx.github.GetFile(ctx.owner, ctx.repo, ctx.versionFile, ctx.releaseHead)
	if err != nil {
		return err
	}

	if !strings.Contains(*file.Content, ctx.releaseVersion) {
		return &util.CheckError{Action: util.CheckErrorRetry, Err: fmt.Errorf("version file does not contain release version")}
	}

	pr, err := ctx.github.GetPR(ctx.owner, ctx.repo, ctx.releaseHead)
	if err != nil {
		return err
	}

	if pr == nil {
		return &util.CheckError{
			Action: util.CheckErrorRetry,
			Err:    fmt.Errorf("PR not found"),
		}
	}

	runs, err := ctx.github.GetCheckRuns(ctx.owner, ctx.repo, ctx.releaseHead)
	if err != nil {
		return err
	}

	for _, run := range runs {
		if run.GetStatus() == "completed" && run.GetConclusion() != "success" {
			return &util.CheckError{
				Action: util.CheckErrorFail,
				Err:    fmt.Errorf("check %s is not successful", run.GetName()),
			}
		}
	}

	for _, run := range runs {
		if run.GetStatus() != "completed" {
			return &util.CheckError{
				Action: util.CheckErrorWait,
				Err:    fmt.Errorf("check %s has not completed yet", run.GetName()),
			}
		}
	}

	return nil
}

func (ctx CutBranch) UpdateVersionFile(head, sha, base, version, title, body string, draft bool) error {
	versionFile, err := ctx.github.GetFile(ctx.owner, ctx.repo, ctx.versionFile, head)
	if err != nil {
		return err
	}

	if !strings.Contains(*versionFile.Content, version) {
		cmd := git.Command{Name: "sed", Args: []string{"-i", fmt.Sprintf("s/const CurrentVersionNumber = \".*\"/const CurrentVersionNumber = \"%s\"/g", version), ctx.versionFile}}
		err = ctx.git.CloneExecCommitAndPush(ctx.owner, ctx.repo, head, sha, ctx.versionFile, fmt.Sprintf("chore: update %s", ctx.versionFile), cmd)
		if err != nil {
			return err
		}
	}

	_, err = ctx.github.GetOrCreatePR(ctx.owner, ctx.repo, head, base, title, body, draft)
	return err
}

func (ctx CutBranch) Run() error {
	releaseHead, err := ctx.github.GetOrCreateBranch(ctx.owner, ctx.repo, ctx.releaseHead, ctx.defaultBase)
	if err != nil {
		return err
	}

	defaultVersionFile, err := ctx.github.GetFile(ctx.owner, ctx.repo, ctx.versionFile, ctx.defaultBase)
	if err != nil {
		return err
	}

	if !strings.Contains(*defaultVersionFile.Content, ctx.defaultVersion) {
		defaultHead, err := ctx.github.GetOrCreateBranch(ctx.owner, ctx.repo, ctx.defaultHead, ctx.defaultBase)
		if err != nil {
			return err
		}

		err = ctx.UpdateVersionFile(ctx.defaultHead, defaultHead.GetCommit().GetSHA(), ctx.defaultBase, ctx.defaultVersion, ctx.defaultTitle, ctx.defaultBody, false)
		if err != nil {
			return err
		}
	}

	return ctx.UpdateVersionFile(ctx.releaseHead, releaseHead.GetCommit().GetSHA(), ctx.releaseBase, ctx.releaseVersion, ctx.releaseTitle, ctx.releaseBody, true)
}

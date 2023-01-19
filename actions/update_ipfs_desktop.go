package actions

import (
	"fmt"
	"strings"

	"github.com/ipfs/kuboreleaser/git"
	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/util"
)

type UpdateIPFSDesktop struct {
	git         *git.Client
	github      *github.Client
	owner       string
	repo        string
	base        string
	head        string
	title       string
	body        string
	version     string
	versionFile string
	glob        string
	draft       bool
}

func NewUpdateIPFSDesktop(git *git.Client, github *github.Client, version *util.Version) (*UpdateIPFSDesktop, error) {
	return &UpdateIPFSDesktop{
		git:         git,
		github:      github,
		owner:       "ipfs",
		repo:        "ipfs-desktop",
		base:        "master",
		head:        fmt.Sprintf("update-kubo-%s", version.MajorMinor()),
		title:       fmt.Sprintf("Update Kubo: %s", version.MajorMinor()),
		body:        fmt.Sprintf("This PR update Kubo to %s", version.MajorMinor()),
		version:     version.Version,
		versionFile: "package.json",
		glob:        "package*.json",
		draft:       version.Prerelease() != "",
	}, nil
}

func (ctx UpdateIPFSDesktop) Check() error {
	file, err := ctx.github.GetFile(ctx.owner, ctx.repo, ctx.versionFile, ctx.base)
	if err != nil {
		return err
	}

	if !strings.Contains(*file.Content, ctx.version) {
		file, err = ctx.github.GetFile(ctx.owner, ctx.repo, ctx.versionFile, ctx.head)
		if err != nil {
			return err
		}

		if !strings.Contains(*file.Content, ctx.version) {
			return &util.CheckError{Action: util.CheckErrorRetry, Err: fmt.Errorf("version file does not contain version")}
		}

		pr, err := ctx.github.GetPR(ctx.owner, ctx.repo, ctx.head)
		if err != nil {
			return err
		}

		if pr == nil {
			return &util.CheckError{
				Action: util.CheckErrorRetry,
				Err:    fmt.Errorf("PR not found"),
			}
		}

		runs, err := ctx.github.GetCheckRuns(ctx.owner, ctx.repo, ctx.head)
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
	}
	return nil
}

func (ctx UpdateIPFSDesktop) Run() error {
	head, err := ctx.github.GetOrCreateBranch(ctx.owner, ctx.repo, ctx.head, ctx.base)
	if err != nil {
		return err
	}

	versionFile, err := ctx.github.GetFile(ctx.owner, ctx.repo, ctx.versionFile, ctx.head)
	if err != nil {
		return err
	}

	if !strings.Contains(*versionFile.Content, ctx.version) {
		cmd := git.Command{Name: "npm", Args: []string{"install", fmt.Sprintf("go-ipfs@%s", ctx.version), "--save"}}
		err = ctx.git.CloneExecCommitAndPush(ctx.owner, ctx.repo, ctx.head, head.GetCommit().GetSHA(), ctx.glob, fmt.Sprintf("chore: update %s", ctx.versionFile), cmd)
		if err != nil {
			return err
		}
	}

	_, err = ctx.github.GetOrCreatePR(ctx.owner, ctx.repo, ctx.head, ctx.base, ctx.title, ctx.body, ctx.draft)
	if err != nil {
		return err
	}

	return nil
}

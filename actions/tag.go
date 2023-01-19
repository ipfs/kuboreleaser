package actions

import (
	"fmt"
	"os"

	"github.com/ipfs/kuboreleaser/git"
	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/util"
)

type Tag struct {
	git     *git.Client
	github  *github.Client
	owner   string
	repo    string
	head    string
	version string
}

func NewTag(git *git.Client, github *github.Client, version *util.Version) (*Tag, error) {
	return &Tag{
		git:     git,
		github:  github,
		owner:   "ipfs",
		repo:    "kubo",
		head:    fmt.Sprintf("release-%s", version.MajorMinor()),
		version: version.Version,
	}, nil
}

func (ctx Tag) Check() error {
	tag, err := ctx.github.GetTag(ctx.owner, ctx.repo, ctx.version)
	if err != nil {
		return err
	}

	if tag == nil {
		return &util.CheckError{Action: util.CheckErrorRetry, Err: fmt.Errorf("tag %s does not exist", ctx.version)}
	}

	return nil
}

func (ctx Tag) Run() error {
	tag, err := ctx.github.GetTag(ctx.owner, ctx.repo, ctx.version)
	if err != nil {
		return err
	}
	if tag != nil {
		return nil
	}

	branch, err := ctx.github.GetBranch(ctx.owner, ctx.repo, ctx.head)
	if err != nil {
		return err
	}

	if branch == nil {
		return fmt.Errorf("branch %s does not exist", ctx.head)
	}

	dir, err := os.MkdirTemp("", "dist")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	sha := branch.GetCommit().GetSHA()

	repo, err := ctx.git.Clone(dir, ctx.owner, ctx.repo, ctx.head, sha)
	if err != nil {
		return err
	}

	ref, err := repo.Tag(sha, ctx.version, fmt.Sprintf("Release %s", ctx.version))
	if err != nil {
		return err
	}

	var confirmation string

	fmt.Printf(`
Tag created:
%v

Signature: %s

The tag will now be pushed to the remote repository.
Only 'yes' will be accepted to approve.

Enter a value: `, ref, ref.PGPSignature)
	fmt.Scanln(&confirmation)

	if confirmation != "yes" {
		return fmt.Errorf("confirmation is not 'yes'")
	}

	err = repo.PushTag(ctx.version)
	if err != nil {
		return err
	}

	return nil
}

package actions

import (
	"bytes"
	"fmt"
	"os"

	gh "github.com/google/go-github/v48/github"
	"github.com/ipfs/kuboreleaser/git"
	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
)

type PrepareBranch struct {
	Git     *git.Client
	GitHub  *github.Client
	Version *util.Version
}

func (ctx PrepareBranch) Check() error {
	versionReleaseBranch := repos.Kubo.VersionReleaseBranch(ctx.Version)

	err := CheckPR(ctx.GitHub, repos.Kubo.Owner, repos.Kubo.Repo, versionReleaseBranch, !ctx.Version.IsPrerelease())
	if err != nil {
		return err
	}

	if !ctx.Version.IsPatch() {
		versionUpdateBranch := repos.Kubo.VersionUpdateBranch(ctx.Version)
		err := CheckPR(ctx.GitHub, repos.Kubo.Owner, repos.Kubo.Repo, versionUpdateBranch, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ctx PrepareBranch) MkReleaseLog() error {
	placeholder := []byte("### 📝 Changelog\n\n### 👨‍👩‍👧‍👦 Contributors\n")
	rootname := "/root/go/src"
	dirname := fmt.Sprintf("%s/github.com/%s/%s", rootname, repos.Kubo.Owner, repos.Kubo.Repo)
	filename := fmt.Sprintf("docs/changelogs/%s.md", ctx.Version.MajorMinor())
	branch := repos.Kubo.VersionReleaseBranch(ctx.Version)

	err := os.Mkdir(rootname, 0755)
	if err != nil {
		return err
	}
	defer os.RemoveAll(rootname)

	cmd := util.Command{
		Name: "git",
		Args: []string{"clone", "https://github.com/ipfs/kubo", dirname},
	}
	err = cmd.Run()
	if err != nil {
		return err
	}

	changelog, err := os.ReadFile(fmt.Sprintf("%s/%s", dirname, filename))
	if err != nil {
		return err
	}
	if !bytes.Contains(changelog, placeholder) {
		return nil
	}

	cmd = util.Command{
		Name: "git",
		Args: []string{"checkout", branch},
		Dir:  dirname,
	}
	err = cmd.Run()
	if err != nil {
		return err
	}

	out := &bytes.Buffer{}
	cmd = util.Command{
		Name: "./mkreleaselog",
		Dir:  dirname,
		Stdout: util.Stdout{
			Writer: out,
		},
	}
	err = cmd.Run()
	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("/root/go/src/github.com/ipfs/kubo/%s", filename), bytes.Replace(changelog, placeholder, out.Bytes(), 1), 0644)
	if err != nil {
		return err
	}

	cmd = util.Command{
		Name: "git",
		Args: []string{"add", filename},
		Dir:  dirname,
	}
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = util.Command{
		Name: "git",
		Args: []string{"commit", "-m", fmt.Sprintf("chore: update changelog for %s", ctx.Version.MajorMinor())},
		Dir:  dirname,
	}
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = util.Command{
		Name: "git",
		Args: []string{"push", "origin", branch},
		Dir:  dirname,
	}
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (ctx PrepareBranch) UpdateVersion(branch, source, currentVersionNumber, base, title, body string, draft bool) (*gh.PullRequest, error) {
	b, err := ctx.GitHub.GetOrCreateBranch(repos.Kubo.Owner, repos.Kubo.Repo, branch, source)
	if err != nil {
		return nil, err
	}

	err = ctx.Git.RunAndPush(repos.Kubo.Owner, repos.Kubo.Repo, branch, b.GetCommit().GetSHA(), "chore: update version", util.Command{Name: "sed", Args: []string{"-i", fmt.Sprintf("s/const CurrentVersionNumber = \".*\"/const CurrentVersionNumber = \"%s\"/g", currentVersionNumber), "version.go"}})
	if err != nil {
		return nil, err
	}

	return ctx.GitHub.GetOrCreatePR(repos.Kubo.Owner, repos.Kubo.Repo, branch, base, title, body, draft)
}

func (ctx PrepareBranch) Run() error {
	dev := fmt.Sprintf("%s.0-dev", ctx.Version.NextMajorMinor())

	branch := repos.Kubo.VersionReleaseBranch(ctx.Version)
	var source string
	if ctx.Version.IsPatch() {
		source = repos.Kubo.ReleaseBranch
	} else {
		source = repos.Kubo.DefaultBranch
	}
	currentVersionNumber := ctx.Version.String()[1:]
	base := repos.Kubo.ReleaseBranch
	title := fmt.Sprintf("Release: %s", ctx.Version.MajorMinorPatch())
	body := fmt.Sprintf("This PR creates release %s", ctx.Version.MajorMinorPatch())
	draft := ctx.Version.IsPrerelease()

	pr, err := ctx.UpdateVersion(branch, source, currentVersionNumber, base, title, body, draft)
	if err != nil {
		return err
	}

	prompt := fmt.Sprintf(`If needed, check out the %s branch of %s/%s repository and cherry-pick commits from %s using the following command:

git cherry-pick -x <commit>

Please approve after all the required commits are cherry-picked.`, branch, repos.Kubo.Owner, repos.Kubo.Repo, repos.Kubo.DefaultBranch)
	if !util.Confirm(prompt) {
		return fmt.Errorf("cherry-picking commits from %s to %s was not confirmed", repos.Kubo.DefaultBranch, branch)
	}

	if !ctx.Version.IsPrerelease() {
		err := ctx.MkReleaseLog()
		if err != nil {
			return err
		}

		if !util.ConfirmPR(pr) {
			return fmt.Errorf("pr not merged")
		}
	}

	if !ctx.Version.IsPatch() {
		branch = repos.Kubo.VersionUpdateBranch(ctx.Version)
		source = repos.Kubo.DefaultBranch
		currentVersionNumber = dev[1:]
		base = repos.Kubo.DefaultBranch
		title = fmt.Sprintf("Update Version: %s", ctx.Version.MajorMinor())
		body = fmt.Sprintf("This PR updates version as part of the %s release", ctx.Version.MajorMinor())
		draft = false

		pr, err := ctx.UpdateVersion(branch, source, currentVersionNumber, base, title, body, draft)
		if err != nil {
			return err
		}
		if !util.ConfirmPR(pr) {
			return fmt.Errorf("pr not merged")
		}
	}

	return nil
}

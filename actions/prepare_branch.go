package actions

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"

	gh "github.com/google/go-github/v48/github"
	"github.com/ipfs/kuboreleaser/git"
	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
	log "github.com/sirupsen/logrus"
)

type PrepareBranch struct {
	Git     *git.Client
	GitHub  *github.Client
	Version *util.Version
}

func (ctx PrepareBranch) Check() error {
	log.Info("I'm going to check if PRs that update versions in the release branch and the master branch exist and if they're merged already.")

	versionReleaseBranch := repos.Kubo.VersionReleaseBranch(ctx.Version)

	err := CheckPR(ctx.GitHub, repos.Kubo.Owner, repos.Kubo.Repo, versionReleaseBranch, !ctx.Version.IsPrerelease())
	if err != nil {
		return err
	}
	// Should we check if the PR checks are passing?

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

	name := util.GetenvPrompt("GITHUB_USER_NAME")
	email := util.GetenvPrompt("GITHUB_USER_EMAIL")
	token := util.GetenvPromptSecret("GITHUB_TOKEN", "The token should have the following scopes: ... Please enter the token:")

	err := os.MkdirAll(rootname, 0755)
	if err != nil {
		return err
	}
	defer os.RemoveAll(rootname)

	cmd := util.Command{
		Name: "git",
		Args: []string{"clone", fmt.Sprintf("https://%s@github.com/ipfs/kubo", token), dirname},
	}
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = util.Command{
		Name: "git",
		Args: []string{"config", "user.name", name},
		Dir:  dirname,
	}
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = util.Command{
		Name: "git",
		Args: []string{"config", "user.email", email},
		Dir:  dirname,
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
		Name: "./bin/mkreleaselog",
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

	pr, err := ctx.GitHub.GetOrCreatePR(repos.Kubo.Owner, repos.Kubo.Repo, branch, base, title, body, draft)
	if err != nil {
		return nil, err
	}
	return pr, nil
}

func (ctx PrepareBranch) GetBody(branch, foreword string) (string, error) {
	kuboCommits, err := ctx.GitHub.Compare(repos.Kubo.Owner, repos.Kubo.Repo, branch, repos.Kubo.DefaultBranch)
	if err != nil {
		return "", err
	}
	file, err := ctx.GitHub.GetFile(repos.Kubo.Owner, repos.Kubo.Repo, "go.mod", branch)
	if err != nil {
		return "", err
	}
	if file == nil {
		return "", fmt.Errorf("🚨 https://github.com/%s/%s/tree/%s/go.mod not found", repos.Kubo.Owner, repos.Kubo.Repo, branch)
	}

	content, err := base64.StdEncoding.DecodeString(*file.Content)
	if err != nil {
		return "", err
	}

	// find the boxo version
	boxoVersion := ""
	for _, line := range strings.Split(string(content[:]), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, fmt.Sprintf("github.com/%s/%s", repos.Boxo.Owner, repos.Boxo.Repo)) {
			boxoVersion = strings.Split(line, " ")[1]
			break
		}
	}
	if boxoVersion == "" {
		return "", fmt.Errorf("🚨 boxo version not found in https://github.com/%s/%s/tree/%s/go.mod", repos.Kubo.Owner, repos.Kubo.Repo, branch)
	}

	// find the boxo commit or tag in boxo version
	boxoBranch := ""
	if strings.Contains(boxoVersion, "-") {
		boxoBranch = strings.Split(boxoVersion, "-")[2]
	} else {
		boxoBranch = boxoVersion
	}

	boxoCommits, err := ctx.GitHub.Compare(repos.Boxo.Owner, repos.Boxo.Repo, boxoBranch, repos.Boxo.DefaultBranch)
	if err != nil {
		return "", err
	}

	kuboCommitsStr := "```\n"
	for _, commit := range kuboCommits {
		kuboCommitsStr += fmt.Sprintf("%s %s\n", commit.GetSHA()[:7], strings.Split(commit.GetCommit().GetMessage(), "\n")[0])
	}
	kuboCommitsStr += "```"

	boxoCommitsStr := "```\n"
	for _, commit := range boxoCommits {
		boxoCommitsStr += fmt.Sprintf("%s %s\n", commit.GetSHA()[:7], strings.Split(commit.GetCommit().GetMessage(), "\n")[0])
	}
	boxoCommitsStr += "```"

	return fmt.Sprintf(`%s

---

#### Kubo commits **NOT** included in this release

%s

#### Boxo commits **NOT** included in this release

%s`, foreword, kuboCommitsStr, boxoCommitsStr), nil
}

func (ctx PrepareBranch) Run() error {
	log.Info("I'm going to create PRs that update the version in the release branch and the master branch.")
	log.Info("I'm also going to update the changelog if we're performing the final release. Please note that it might take a while because I have to clone a looooot of repos.")

	dev := fmt.Sprintf("%s.0-dev", ctx.Version.NextMajorMinor())

	branch := repos.Kubo.VersionReleaseBranch(ctx.Version)
	var source string
	if ctx.Version.IsPatch() {
		// NOTE: For patch releases we want to create the new release branch from the previous release branch, e.g.
		// when creating release-0.50.6, we want to create it from release-0.50.5
		patchVersion, err := strconv.Atoi(ctx.Version.Patch())
		if err != nil {
			return err
		}
		previousVersionString := fmt.Sprintf("%s.%s", ctx.Version.MajorMinor(), strconv.Itoa(patchVersion-1))
		previousVersion, err := util.NewVersion(previousVersionString)
		if err != nil {
			return err
		}
		source = repos.Kubo.VersionReleaseBranch(previousVersion)
	} else {
		source = repos.Kubo.DefaultBranch
	}
	currentVersionNumber := ctx.Version.String()[1:]
	base := repos.Kubo.ReleaseBranch
	title := fmt.Sprintf("Release: %s [skip changelog]", ctx.Version.MajorMinorPatch())
	body := fmt.Sprintf("This PR creates release %s", ctx.Version.MajorMinorPatch())
	draft := ctx.Version.IsPrerelease()

	// NOTE: This should update const CurrentVersionNumber in version.go to the full version without a v prefix
	// on the version release branch created from source
	pr, err := ctx.UpdateVersion(branch, source, currentVersionNumber, base, title, body, draft)
	if err != nil {
		return err
	}

	body, err = ctx.GetBody(branch, body)
	if err != nil {
		return err
	}

	pr.Body = &body
	err = ctx.GitHub.UpdatePR(pr)
	if err != nil {
		return err
	}

	fmt.Printf("💁 Your release PR is ready at %s\n", pr.GetHTMLURL())

	// TODO: check for conflicts and tell the user to resolve them
	// or resolve them automatically with git merge origin/release -X ours

	prompt := fmt.Sprintf(`If needed, check out the %s branch of %s/%s repository and cherry-pick commits from %s using the following command:

git cherry-pick -x <commit>

Please approve after all the required commits are cherry-picked.`, branch, repos.Kubo.Owner, repos.Kubo.Repo, repos.Kubo.DefaultBranch)
	if !util.Confirm(prompt) {
		return fmt.Errorf("🚨 cherry-picking commits to https://github.com/%s/%s/tree/%s was not confirmed correctly", repos.Kubo.Owner, repos.Kubo.Repo, branch)
	}

	if !ctx.Version.IsPrerelease() {
		err := ctx.MkReleaseLog()
		if err != nil {
			return err
		}

		fmt.Println("Use merge commit to merge this PR! You'll have to tag it after the merge.")
		if !util.ConfirmPR(pr) {
			return fmt.Errorf("🚨 %s not merged", pr.GetHTMLURL())
		}
	}

	if !ctx.Version.IsPatch() {
		branch = repos.Kubo.VersionUpdateBranch(ctx.Version)
		source = repos.Kubo.DefaultBranch
		currentVersionNumber = dev[1:]
		base = repos.Kubo.DefaultBranch
		title = fmt.Sprintf("Update Version: %s [skip changelog]", ctx.Version.MajorMinor())
		body = fmt.Sprintf("This PR updates version as part of the %s release", ctx.Version.MajorMinor())
		draft = false

		pr, err := ctx.UpdateVersion(branch, source, currentVersionNumber, base, title, body, draft)
		if err != nil {
			return err
		}

		if ctx.Version.IsPrerelease() {
			fmt.Printf(`💁 Release PR ready at %s. Do not merge it.`, pr.GetHTMLURL())
		} else if !pr.GetMerged() && !util.ConfirmPR(pr) {
			return fmt.Errorf("🚨 %s not merged", pr.GetHTMLURL())
		}
	}

	return nil
}

package actions

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/matrix"
	"github.com/ipfs/kuboreleaser/repos"
	"github.com/ipfs/kuboreleaser/util"
)

type Promote struct {
	GitHub  *github.Client
	Matrix  *matrix.Client
	Version *util.Version
}

func (ctx *Promote) getDiscoursePostTitle() string {
	return fmt.Sprintf("Kubo %s is out!", ctx.Version)
}

func (ctx *Promote) getDiscoursePostBody() string {
	return fmt.Sprintf(`## Kubo %s is out!

See:
- Code: https://github.com/ipfs/kubo/releases/tag/%s
- Binaries: https://dist.ipfs.tech/kubo/%s/
- Docker: `+"`docker pull ipfs/kubo:%s`"+`
- Release Notes (WIP): https://github.com/ipfs/kubo/blob/release-%s/docs/changelogs/%s.md`, ctx.Version, ctx.Version, ctx.Version, ctx.Version, ctx.Version.MajorMinor(), ctx.Version.MajorMinor())
}

func (ctx *Promote) getReleaseIssueComment() string {
	if ctx.Version.IsPrerelease() {
		return fmt.Sprintf(`Early testers ping for %s testing 😄.

	- [ ] pacman.store (@RubenKelevra)
	- [ ] Infura (@MichaelMure)
	- [ ] Textile (@sanderpick)
	- [ ] Pinata (@obo20)
	- [ ] RTrade (@postables)
	- [ ] QRI (@b5)
	- [ ] Siderus (@koalalorenzo)
	- [ ] Charity Engine (@rytiss, @tristanolive)
	- [ ] Fission (@bmann)
	- [ ] OrbitDB (@aphelionz)

	You're getting this message because you're listed [here](https://github.com/ipfs/kubo/blob/master/docs/EARLY_TESTERS.md#who-has-signed-up). Please update this list if you no longer want to be included.`, ctx.Version)
	} else {
		return fmt.Sprintf("🎉 Kubo [%s](https://github.com/ipfs/kubo/releases/tag/%s) is out!", ctx.Version, ctx.Version)
	}
}

func (ctx Promote) Check() error {
	issue, err := ctx.GitHub.GetIssue(repos.Kubo.Owner, repos.Kubo.Repo, repos.Kubo.ReleaseIssueTitle(ctx.Version))
	if err != nil {
		return err
	}
	if issue == nil {
		return fmt.Errorf("issue %s not found (%w)", repos.Kubo.ReleaseIssueTitle(ctx.Version), ErrFailure)
	}

	comment, err := ctx.GitHub.GetIssueComment(repos.Kubo.Owner, repos.Kubo.Repo, issue.GetNumber(), ctx.getReleaseIssueComment())
	if err != nil {
		return err
	}
	if comment == nil {
		return fmt.Errorf("comment %s not found (%w)", ctx.getReleaseIssueComment(), ErrIncomplete)
	}

	messages, err := ctx.Matrix.GetLatestMessagesBy("#ipfs-chatter:ipfs.io", "@ipfsbot:matrix.org", 10)
	if err != nil {
		return err
	}

	var found bool
	for _, message := range messages {
		body, ok := message.Body()
		if ok && strings.Contains(body, ctx.getDiscoursePostTitle()) {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("post %s not found (%w)", ctx.getDiscoursePostTitle(), ErrIncomplete)
	}

	if !ctx.Version.IsPrerelease() {
		release, err := ctx.GitHub.GetRelease(repos.Kubo.Owner, repos.Kubo.Repo, ctx.Version.String())
		if err != nil {
			return err
		}
		if release == nil {
			return fmt.Errorf("release %s not found (%w)", ctx.Version, ErrFailure)
		}
		if !strings.Contains(release.GetBody(), "- 💬 [Discuss]") {
			return fmt.Errorf("release %s does not contain a discuss link (%w)", ctx.Version, ErrIncomplete)
		}
	}

	return nil
}

func (ctx Promote) Run() error {
	url := repos.Kubo.ReleaseURL(ctx.Version)

	issue, err := ctx.GitHub.GetIssue(repos.Kubo.Owner, repos.Kubo.Repo, repos.Kubo.ReleaseIssueTitle(ctx.Version))
	if err != nil {
		return err
	}
	if issue == nil {
		return fmt.Errorf("issue not found")
	}

	_, err = ctx.GitHub.GetOrCreateIssueComment(repos.Kubo.Owner, repos.Kubo.Repo, issue.GetNumber(), ctx.getReleaseIssueComment())
	if err != nil {
		return err
	}

	prompt := fmt.Sprintf(`IPFS Discourse does not have API access enabled.

Please go to https://discuss.ipfs.io and create a new topic with the following content:
Title: %s
Category: News
Tags: kubo, go-ipfs
Body: %s

Remember to pin the topic globally!`, ctx.getDiscoursePostTitle(), ctx.getDiscoursePostBody())
	if !util.Confirm(prompt) {
		return fmt.Errorf("discourse post not created")
	}

	if !ctx.Version.IsPrerelease() {
		prompt := fmt.Sprintf(`Go to %s and add the link to the IPFS Discourse post to the top of the release notes.

Use the following template:
- 💬 [Discuss](https://discuss.ipfs.io/t/kubo-%s-is-out/XXXX)`, url, strings.ReplaceAll(ctx.Version.String(), ".", "-"))

		if !util.Confirm(prompt) {
			return fmt.Errorf("discourse post not added to release notes")
		}
	}

	if !ctx.Version.IsPrerelease() && !ctx.Version.IsPatch() {
		prompt := fmt.Sprintf(`Reddit supports only OAuth2 authentication.

Please go to https://www.reddit.com/r/ipfs/new/ and create a new "Link" post with the following content:

Url: %s`, url)
		if !util.Confirm(prompt) {
			return fmt.Errorf("reddit post not created")
		}

		file, err := ctx.GitHub.GetFile(repos.Kubo.Owner, repos.Kubo.Repo, "docs/changelogs/"+ctx.Version.MajorMinor()+".md", "release")
		if err != nil {
			return err
		}
		if file == nil {
			return fmt.Errorf("changelog not found")
		}

		content, err := base64.StdEncoding.DecodeString(*file.Content)
		if err != nil {
			return err
		}

		highlights := []string{}
		for _, line := range strings.Split(string(content[:]), "\n") {
			if strings.HasPrefix(line, "##### ") {
				highlights = append(highlights, line[6:])
			}
		}

		prompt = fmt.Sprintf(`We do not have direct access to the IPFS Twitter account.

Please go to https://filecoinproject.slack.com/archives/C018EJ8LWH1 (#shared-pl-marketing-requests in FIL Slack) and ask the team to create a new tweet with the following content:

What's happening?: #Kubo %s was just released!
%s
%s`, ctx.Version, strings.Join(highlights, "\n"), url)
		if !util.Confirm(prompt) {
			return fmt.Errorf("twitter post not created")
		}
	}

	return nil
}

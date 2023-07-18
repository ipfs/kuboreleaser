package github

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/shurcooL/githubv4"
	log "github.com/sirupsen/logrus"

	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
)

type Client struct {
	v3 *github.Client
	v4 *githubv4.Client
}

func NewClient() (*Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN not set")
	}

	sts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	o2 := oauth2.NewClient(context.Background(), sts)

	return &Client{
		v3: github.NewClient(o2),
		v4: githubv4.NewClient(o2),
	}, nil
}

func (c *Client) GetIssue(owner, repo, title string) (*github.Issue, error) {
	log.WithFields(log.Fields{
		"owner": owner,
		"repo":  repo,
		"title": title,
	}).Debug("Searching for issue...")

	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	q := fmt.Sprintf("is:issue repo:%s/%s in:title %s", owner, repo, title)
	var issue *github.Issue
	for {
		is, r, err := c.v3.Search.Issues(context.Background(), q, opt)
		if err != nil {
			return nil, err
		}
		for _, i := range is.Issues {
			if i.GetTitle() == title {
				issue = i
				break
			}
		}
		if issue != nil || r.NextPage == 0 {
			break
		}
		opt.Page = r.NextPage
	}

	if issue != nil {
		log.WithFields(log.Fields{
			"url": issue.GetHTMLURL(),
		}).Debug("Found issue")
	} else {
		log.Debug("Issue not found")
	}

	return issue, nil
}

func (c *Client) CreateIssue(owner, repo, title, body string) (*github.Issue, error) {
	log.WithFields(log.Fields{
		"owner": owner,
		"repo":  repo,
		"title": title,
		"body":  body,
	}).Debug("Creating issue...")

	issue, _, err := c.v3.Issues.Create(context.Background(), owner, repo, &github.IssueRequest{
		Title: &title,
		Body:  &body,
	})

	if issue != nil {
		log.WithFields(log.Fields{
			"url": issue.GetHTMLURL(),
		}).Debug("Created issue")
	} else {
		log.Debug("Issue not created")
	}

	return issue, err
}

func (c *Client) GetOrCreateIssue(owner, repo, title, body string) (*github.Issue, error) {
	issue, err := c.GetIssue(owner, repo, title)
	if err != nil {
		return nil, err
	}
	if issue != nil {
		return issue, nil
	}
	return c.CreateIssue(owner, repo, title, body)
}

func (c *Client) GetIssueComment(owner, repo string, number int, body string) (*github.IssueComment, error) {
	log.WithFields(log.Fields{
		"owner":  owner,
		"repo":   repo,
		"number": number,
		"body":   body,
	}).Debug("Searching for issue comment...")

	opt := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var comment *github.IssueComment
	for {
		cs, r, err := c.v3.Issues.ListComments(context.Background(), owner, repo, number, opt)
		if err != nil {
			return nil, err
		}
		for _, c := range cs {
			if c.GetBody() == body {
				comment = c
				break
			}
		}
		if comment != nil || r.NextPage == 0 {
			break
		}
		opt.Page = r.NextPage
	}

	if comment != nil {
		log.WithFields(log.Fields{
			"url": comment.GetHTMLURL(),
		}).Debug("Found comment")
	} else {
		log.Debug("Comment not found")
	}

	return comment, nil
}

func (c *Client) CreateIssueComment(owner, repo string, number int, body string) (*github.IssueComment, error) {
	log.WithFields(log.Fields{
		"owner":  owner,
		"repo":   repo,
		"number": number,
		"body":   body,
	}).Debug("Creating issue comment...")

	comment, _, err := c.v3.Issues.CreateComment(context.Background(), owner, repo, number, &github.IssueComment{
		Body: &body,
	})

	if comment != nil {
		log.WithFields(log.Fields{
			"url": comment.GetHTMLURL(),
		}).Debug("Created comment")
	} else {
		log.Debug("Comment not created")
	}

	return comment, err
}

func (c *Client) GetOrCreateIssueComment(owner, repo string, number int, body string) (*github.IssueComment, error) {
	comment, err := c.GetIssueComment(owner, repo, number, body)
	if err != nil {
		return nil, err
	}
	if comment != nil {
		return comment, nil
	}
	return c.CreateIssueComment(owner, repo, number, body)
}

func (c *Client) GetBranch(owner, repo, name string) (*github.Branch, error) {
	log.WithFields(log.Fields{
		"owner": owner,
		"repo":  repo,
		"name":  name,
	}).Debug("Searching for branch...")

	branch, _, err := c.v3.Repositories.GetBranch(context.Background(), owner, repo, name, false)
	if err != nil && strings.Contains(err.Error(), "404") {
		return nil, nil
	}

	if branch != nil {
		log.WithFields(log.Fields{
			"commit": branch.GetCommit().GetSHA(),
		}).Debug("Found branch")
	} else {
		log.Debug("Branch not found")
	}

	return branch, err
}

func (c *Client) CreateBranch(owner, repo, name, source string) (*github.Branch, error) {
	log.WithFields(log.Fields{
		"owner":  owner,
		"repo":   repo,
		"name":   name,
		"source": source,
	}).Debug("Creating branch...")

	r, _, err := c.v3.Git.GetRef(context.Background(), owner, repo, "refs/heads/"+source)
	if err != nil {
		return nil, err
	}

	b, _, err := c.v3.Git.CreateRef(context.Background(), owner, repo, &github.Reference{
		Ref:    github.String("refs/heads/" + name),
		Object: r.GetObject(),
	})
	if err != nil {
		return nil, err
	}

	if b != nil {
		log.WithFields(log.Fields{
			"url": b.GetURL(),
		}).Debug("Created branch")
	} else {
		log.Debug("Branch not created")
	}

	return c.GetBranch(owner, repo, name)
}

func (c *Client) GetOrCreateBranch(owner, repo, name, source string) (*github.Branch, error) {
	branch, err := c.GetBranch(owner, repo, name)
	if err != nil {
		return nil, err
	}
	if branch != nil {
		return branch, nil
	}

	return c.CreateBranch(owner, repo, name, source)
}

func (c *Client) GetPR(owner, repo, head string) (*github.PullRequest, error) {
	log.WithFields(log.Fields{
		"owner": owner,
		"repo":  repo,
		"head":  head,
	}).Debug("Searching for PR...")

	q := fmt.Sprintf("is:pr repo:%s/%s head:%s", owner, repo, head)
	r, _, err := c.v3.Search.Issues(context.Background(), q, &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	})
	if err != nil {
		return nil, err
	}
	if len(r.Issues) == 0 {
		return nil, nil
	}
	for _, i := range r.Issues {
		n := i.GetNumber()
		pr, _, err := c.v3.PullRequests.Get(context.Background(), owner, repo, n)
		if err != nil {
			return nil, err
		}
		if pr != nil {
			if pr.GetHead().GetRef() == head {
				log.WithFields(log.Fields{
					"url": pr.GetHTMLURL(),
				}).Debug("Found PR")
				return pr, nil
			}
		}
	}

	log.Debug("PR not found")
	return nil, nil
}

func (c *Client) CreatePR(owner, repo, head, base, title, body string, draft bool) (*github.PullRequest, error) {
	log.WithFields(log.Fields{
		"owner": owner,
		"repo":  repo,
		"head":  head,
		"base":  base,
		"title": title,
		"body":  body,
		"draft": draft,
	}).Debug("Creating PR...")

	pr, _, err := c.v3.PullRequests.Create(context.Background(), owner, repo, &github.NewPullRequest{
		Title: &title,
		Head:  &head,
		Base:  &base,
		Body:  &body,
		Draft: &draft,
	})

	if pr != nil {
		log.WithFields(log.Fields{
			"url": pr.GetHTMLURL(),
		}).Debug("Created PR")
	} else {
		log.Debug("PR not created")
	}

	return pr, err
}

func (c *Client) GetOrCreatePR(owner, repo, head, base, title, body string, draft bool) (*github.PullRequest, error) {
	pr, err := c.GetPR(owner, repo, head)
	if err != nil {
		return nil, err
	}
	if pr != nil && pr.GetMerged() {
		return pr, nil
	}
	if pr == nil || pr.GetState() == "closed" {
		pr, err = c.CreatePR(owner, repo, head, base, title, body, draft)
		if err != nil {
			return nil, err
		}
	}
	if !draft && pr.GetDraft() {
		var m struct {
			MarkPullRequestReadyForReview struct {
				PullRequest struct {
					ID githubv4.ID
				}
			} `graphql:"markPullRequestReadyForReview(input: $input)"`
		}
		input := githubv4.MarkPullRequestReadyForReviewInput{
			PullRequestID: pr.GetNodeID(),
		}
		err = c.v4.Mutate(context.Background(), &m, input, nil)
		if err != nil {
			return pr, err
		}
		pr.Draft = &draft
	}
	return pr, nil
}

func (c *Client) UpdatePR(pr *github.PullRequest) error {
	log.WithFields(log.Fields{
		"owner":  pr.Base.Repo.Owner.GetLogin(),
		"repo":   pr.Base.Repo.GetName(),
		"number": pr.GetNumber(),
	}).Debug("Updating PR...")

	_, _, err := c.v3.PullRequests.Edit(context.Background(), pr.Base.Repo.Owner.GetLogin(), pr.Base.Repo.GetName(), pr.GetNumber(), pr)
	return err
}

func (c *Client) GetFile(owner, repo, path, ref string) (*github.RepositoryContent, error) {
	log.WithFields(log.Fields{
		"owner": owner,
		"repo":  repo,
		"path":  path,
		"ref":   ref,
	}).Debug("Searching for file...")

	f, _, _, err := c.v3.Repositories.GetContents(context.Background(), owner, repo, path, &github.RepositoryContentGetOptions{
		Ref: ref,
	})

	if f != nil {
		log.WithFields(log.Fields{
			"url": f.GetHTMLURL(),
		}).Debug("Found file")
	} else {
		log.Debug("File not found")
	}

	if err != nil && strings.Contains(err.Error(), "404") {
		return nil, nil
	}
	return f, err
}

func (c *Client) GetCheckRuns(owner, repo, ref string) ([]*github.CheckRun, error) {
	log.WithFields(log.Fields{
		"owner": owner,
		"repo":  repo,
		"ref":   ref,
	}).Debug("Searching for check runs...")

	opt := &github.ListCheckRunsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var runs []*github.CheckRun
	for {
		rs, r, err := c.v3.Checks.ListCheckRunsForRef(context.Background(), owner, repo, ref, opt)
		if err != nil {
			return nil, err
		}
		runs = append(runs, rs.CheckRuns...)
		if r.NextPage == 0 {
			break
		}
		opt.Page = r.NextPage
	}

	if len(runs) > 0 {
		urls := []string{}
		for _, r := range runs {
			urls = append(urls, r.GetHTMLURL())
		}
		log.WithFields(log.Fields{
			"urls": urls,
		}).Debug("Found check runs")
	} else {
		log.Debug("Check runs not found")
	}

	return runs, nil
}

func (c *Client) GetIncompleteCheckRuns(owner, repo, ref string) ([]*github.CheckRun, error) {
	log.WithFields(log.Fields{
		"owner": owner,
		"repo":  repo,
		"ref":   ref,
	}).Debug("Searching for incomplete check runs...")

	runs, err := c.GetCheckRuns(owner, repo, ref)
	if err != nil {
		return nil, err
	}

	var incomplete []*github.CheckRun
	for _, r := range runs {
		if r.GetStatus() != "completed" {
			incomplete = append(incomplete, r)
		}
	}

	if len(incomplete) > 0 {
		urls := []string{}
		for _, r := range incomplete {
			urls = append(urls, r.GetHTMLURL())
		}
		log.WithFields(log.Fields{
			"urls": urls,
		}).Debug("Found incomplete check runs")
	} else {
		log.Debug("Incomplete check runs not found")
	}

	return incomplete, nil
}

func (c *Client) GetUnsuccessfulCheckRuns(owner, repo, ref string) ([]*github.CheckRun, error) {
	log.WithFields(log.Fields{
		"owner": owner,
		"repo":  repo,
		"ref":   ref,
	}).Debug("Searching for unsuccessful check runs...")

	runs, err := c.GetCheckRuns(owner, repo, ref)
	if err != nil {
		return nil, err
	}

	var unsuccessful []*github.CheckRun
	for _, r := range runs {
		if r.GetStatus() == "completed" && r.GetConclusion() != "success" && r.GetConclusion() != "skipped" {
			unsuccessful = append(unsuccessful, r)
		}
	}

	if len(unsuccessful) > 0 {
		urls := []string{}
		for _, r := range unsuccessful {
			urls = append(urls, r.GetHTMLURL())
		}
		log.WithFields(log.Fields{
			"urls": urls,
		}).Debug("Found unsuccessful check runs")
	} else {
		log.Debug("Unsuccessful check runs not found")
	}

	return unsuccessful, nil
}

type WorkflowRunInput struct {
	Name  string
	Value interface{}
}

func (c *Client) CreateWorkflowRun(owner, repo, file, ref string, inputs ...WorkflowRunInput) error {
	log.WithFields(log.Fields{
		"owner":  owner,
		"repo":   repo,
		"file":   file,
		"ref":    ref,
		"inputs": inputs,
	}).Debug("Creating workflow run...")

	is := make(map[string]interface{})
	for _, i := range inputs {
		is[i.Name] = i.Value
	}

	_, err := c.v3.Actions.CreateWorkflowDispatchEventByFileName(context.Background(), owner, repo, file, github.CreateWorkflowDispatchEventRequest{
		Ref:    ref,
		Inputs: is,
	})

	if err != nil {
		log.Debug("Failed to create workflow run")
	} else {
		log.Debug("Created workflow run")
	}

	return err
}

func (c *Client) GetWorkflowRun(owner, repo, branch, file string, completed bool) (*github.WorkflowRun, error) {
	log.WithFields(log.Fields{
		"owner":     owner,
		"repo":      repo,
		"branch":    branch,
		"file":      file,
		"completed": completed,
	}).Debug("Searching for workflow run...")

	opt := &github.ListWorkflowRunsOptions{
		Branch:      branch,
		ListOptions: github.ListOptions{PerPage: 1},
	}
	if completed {
		opt.Status = "completed"
	}
	r, _, err := c.v3.Actions.ListWorkflowRunsByFileName(context.Background(), owner, repo, file, opt)
	if err != nil {
		return nil, err
	}

	if len(r.WorkflowRuns) > 0 {
		log.WithFields(log.Fields{
			"url": r.WorkflowRuns[0].GetHTMLURL(),
		}).Debug("Found workflow run")
		return r.WorkflowRuns[0], nil
	} else {
		log.Debug("Workflow run not found")
		return nil, nil
	}
}

type WorkflowRunJobStepLogs struct {
	Name    string
	RawLogs string
}

type WorkflowRunJobLogs struct {
	Name        string
	RawLogs     string
	JobStepLogs map[string]*WorkflowRunJobStepLogs
}

type WorkflowRunLogs struct {
	JobLogs map[string]*WorkflowRunJobLogs
}

func (c *Client) GetWorkflowRunLogs(owner, repo string, id int64) (*WorkflowRunLogs, error) {
	log.WithFields(log.Fields{
		"owner": owner,
		"repo":  repo,
		"id":    id,
	}).Debug("Searching for workflow run logs...")

	url, _, err := c.v3.Actions.GetWorkflowRunLogs(context.Background(), owner, repo, id, true)
	if err != nil {
		return nil, err
	}

	r, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	reader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, err
	}

	logs := &WorkflowRunLogs{
		JobLogs: make(map[string]*WorkflowRunJobLogs),
	}
	sort.Slice(reader.File, func(i, j int) bool {
		return reader.File[i].Name < reader.File[j].Name
	})
	for _, f := range reader.File {
		if !f.FileInfo().IsDir() {
			name := strings.TrimSuffix(f.Name, ".txt")
			if strings.HasSuffix(name, ")") {
				name = name[:strings.LastIndex(name, "(")-1]
			}
			c, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer c.Close()
			b, err := io.ReadAll(c)
			if err != nil {
				return nil, err
			}
			job, step, _ := strings.Cut(name, "/")
			if step == "" {
				_, job, _ = strings.Cut(job, "_")
			} else {
				_, step, _ = strings.Cut(step, "_")
			}
			j := logs.JobLogs[job]
			if j == nil {
				j = &WorkflowRunJobLogs{
					Name:        job,
					JobStepLogs: make(map[string]*WorkflowRunJobStepLogs),
				}
				logs.JobLogs[job] = j
			}
			if step == "" {
				j.RawLogs += string(b)
			} else {
				s := j.JobStepLogs[step]
				if s == nil {
					s = &WorkflowRunJobStepLogs{
						Name: step,
					}
					j.JobStepLogs[step] = s
				}
				s.RawLogs += string(b)
			}
		}
	}

	log.WithFields(log.Fields{
		"url": url,
	}).Debug("Got workflow run logs")

	return logs, nil
}

func (c *Client) GetRelease(owner, repo, tag string) (*github.RepositoryRelease, error) {
	log.WithFields(log.Fields{
		"owner": owner,
		"repo":  repo,
		"tag":   tag,
	}).Debug("Searching for release...")

	r, _, err := c.v3.Repositories.GetReleaseByTag(context.Background(), owner, repo, tag)
	if err != nil && strings.Contains(err.Error(), "404") {
		return nil, nil
	}

	if r != nil {
		log.WithFields(log.Fields{
			"url": r.GetHTMLURL(),
		}).Debug("Found release")
	} else {
		log.Debug("Release not found")
	}

	return r, err
}

func (c *Client) CreateRelease(owner, repo, tag, name, body string, prerelease bool) (*github.RepositoryRelease, error) {
	log.WithFields(log.Fields{
		"owner":      owner,
		"repo":       repo,
		"tag":        tag,
		"name":       name,
		"body":       body,
		"prerelease": prerelease,
	}).Debug("Creating release...")

	r, _, err := c.v3.Repositories.CreateRelease(context.Background(), owner, repo, &github.RepositoryRelease{
		TagName:    &tag,
		Name:       &name,
		Body:       &body,
		Prerelease: &prerelease,
	})

	if r != nil {
		log.WithFields(log.Fields{
			"url": r.GetHTMLURL(),
		}).Debug("Created release")
	} else {
		log.Debug("Release not created")
	}

	return r, err
}

func (c *Client) GetOrCreateRelease(owner, repo, tag, name, body string, prerelease bool) (*github.RepositoryRelease, error) {
	r, err := c.GetRelease(owner, repo, tag)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return c.CreateRelease(owner, repo, tag, name, body, prerelease)
	}
	return r, nil
}

func (c *Client) GetTag(owner, repo, tag string) (*github.Tag, error) {
	log.WithFields(log.Fields{
		"owner": owner,
		"repo":  repo,
		"tag":   tag,
	}).Debug("Searching for tag...")

	r, _, err := c.v3.Git.GetRef(context.Background(), owner, repo, fmt.Sprintf("tags/%s", tag))
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return nil, nil
		}
		return nil, err
	}

	t, _, err := c.v3.Git.GetTag(context.Background(), owner, repo, r.Object.GetSHA())
	if err != nil && strings.Contains(err.Error(), "404") {
		return nil, nil
	}

	if t != nil {
		log.WithFields(log.Fields{
			"url": t.GetURL(),
		}).Debug("Found tag")
	} else {
		log.Debug("Tag not found")
	}

	return t, err
}

func (c *Client) Compare(owner, repo, base, head string) ([]*github.RepositoryCommit, error) {
	log.WithFields(log.Fields{
		"owner": owner,
		"repo":  repo,
		"base":  base,
		"head":  head,
	}).Debug("Comparing...")

	opts := &github.ListOptions{PerPage: 100}
	var commits []*github.RepositoryCommit
	for {
		cs, r, err := c.v3.Repositories.CompareCommits(context.Background(), owner, repo, base, head, opts)
		if err != nil {
			return nil, err
		}
		commits = append(commits, cs.Commits...)
		if r.NextPage == 0 {
			break
		}
		opts.Page = r.NextPage
	}

	if len(commits) > 0 {
		log.WithFields(log.Fields{
			"commits": len(commits),
		}).Debug("Found commits")
	} else {
		log.Debug("Commits not found")
	}

	return commits, nil
}

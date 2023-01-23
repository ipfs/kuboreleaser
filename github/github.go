package github

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
)

type Client struct {
	github *github.Client
}

func NewClient() (*Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("env var GITHUB_TOKEN must be set")
	}

	sts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	o2 := oauth2.NewClient(context.Background(), sts)

	return &Client{
		github: github.NewClient(o2),
	}, nil
}

func (c *Client) GetIssue(owner, repo, title string) (*github.Issue, error) {
	log.Printf("Getting issue [owner: %s, repo: %s, title: %s]\n", owner, repo, title)

	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	q := fmt.Sprintf("is:issue repo:%s/%s in:title %s", owner, repo, title)
	var issue *github.Issue
	for {
		is, r, err := c.github.Search.Issues(context.Background(), q, opt)
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

	return issue, nil
}

func (c *Client) CreateIssue(owner, repo, title, body string) (*github.Issue, error) {
	log.Printf("Creating issue [owner: %s, repo: %s, title: %s]\n", owner, repo, title)

	issue, _, err := c.github.Issues.Create(context.Background(), owner, repo, &github.IssueRequest{
		Title: &title,
		Body:  &body,
	})
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
	log.Printf("Getting issue comment [owner: %s, repo: %s, number: %d, body: %s]\n", owner, repo, number, body)

	opt := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var comment *github.IssueComment
	for {
		cs, r, err := c.github.Issues.ListComments(context.Background(), owner, repo, number, opt)
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

	return comment, nil
}

func (c *Client) CreateIssueComment(owner, repo string, number int, body string) (*github.IssueComment, error) {
	log.Printf("Creating issue comment [owner: %s, repo: %s, number: %d, body: %s]\n", owner, repo, number, body)

	comment, _, err := c.github.Issues.CreateComment(context.Background(), owner, repo, number, &github.IssueComment{
		Body: &body,
	})
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
	log.Printf("Getting branch [owner: %s, repo: %s, name: %s]\n", owner, repo, name)

	branch, _, err := c.github.Repositories.GetBranch(context.Background(), owner, repo, name, false)
	if err != nil && strings.Contains(err.Error(), "404") {
		return nil, nil
	}
	return branch, err
}

func (c *Client) CreateBranch(owner, repo, name, source string) (*github.Branch, error) {
	log.Printf("Creating branch [owner: %s, repo: %s, name: %s, source: %s]\n", owner, repo, name, source)

	r, _, err := c.github.Git.GetRef(context.Background(), owner, repo, "refs/heads/"+source)
	if err != nil {
		return nil, err
	}

	_, _, err = c.github.Git.CreateRef(context.Background(), owner, repo, &github.Reference{
		Ref:    github.String("refs/heads/" + name),
		Object: r.GetObject(),
	})
	if err != nil {
		return nil, err
	}

	return c.GetBranch(owner, repo, name)
}

func (c *Client) GetOrCreateBranch(owner, repo, name, source string) (*github.Branch, error) {
	log.Printf("Getting or creating branch [owner: %s, repo: %s, name: %s, source: %s]\n", owner, repo, name, source)

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
	log.Printf("Getting PR [owner: %s, repo: %s, head: %s]\n", owner, repo, head)

	q := fmt.Sprintf("is:pr repo:%s/%s head:%s", owner, repo, head)
	r, _, err := c.github.Search.Issues(context.Background(), q, &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 1},
	})
	if err != nil {
		return nil, err
	}
	if len(r.Issues) == 0 {
		return nil, nil
	}

	n := r.Issues[0].GetNumber()

	pr, _, err := c.github.PullRequests.Get(context.Background(), owner, repo, n)
	return pr, err
}

func (c *Client) CreatePR(owner, repo, head, base, title, body string, draft bool) (*github.PullRequest, error) {
	log.Printf("Creating PR [owner: %s, repo: %s, head: %s, base: %s, title: %s, body: %s, draft: %t]\n", owner, repo, head, base, title, body, draft)

	pr, _, err := c.github.PullRequests.Create(context.Background(), owner, repo, &github.NewPullRequest{
		Title: &title,
		Head:  &head,
		Base:  &base,
		Body:  &body,
		Draft: &draft,
	})
	return pr, err
}

func (c *Client) GetOrCreatePR(owner, repo, head, base, title, body string, draft bool) (*github.PullRequest, error) {
	log.Printf("Getting or creating PR [owner: %s, repo: %s, head: %s, base: %s, title: %s, body: %s, draft: %t]\n", owner, repo, head, base, title, body, draft)

	pr, err := c.GetPR(owner, repo, head)
	if err != nil {
		return nil, err
	}
	if pr != nil {
		return pr, nil
	}

	return c.CreatePR(owner, repo, head, base, title, body, draft)
}

func (c *Client) GetFile(owner, repo, path, ref string) (*github.RepositoryContent, error) {
	log.Printf("Getting file [owner: %s, repo: %s, path: %s, ref: %s]\n", owner, repo, path, ref)

	f, _, _, err := c.github.Repositories.GetContents(context.Background(), owner, repo, path, &github.RepositoryContentGetOptions{
		Ref: ref,
	})
	if err != nil && strings.Contains(err.Error(), "404") {
		return nil, nil
	}
	return f, err
}

func (c *Client) GetCheckRuns(owner, repo, ref string) ([]*github.CheckRun, error) {
	log.Printf("Getting checks [owner: %s, repo: %s, ref: %s]\n", owner, repo, ref)

	opt := &github.ListCheckRunsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var runs []*github.CheckRun
	for {
		rs, r, err := c.github.Checks.ListCheckRunsForRef(context.Background(), owner, repo, ref, opt)
		if err != nil {
			return nil, err
		}
		runs = append(runs, rs.CheckRuns...)
		if r.NextPage == 0 {
			break
		}
		opt.Page = r.NextPage
	}
	return runs, nil
}

type WorkflowRunInput struct {
	Name  string
	Value interface{}
}

func (c *Client) CreateWorkflowRun(owner, repo, file, ref string, inputs ...WorkflowRunInput) error {
	log.Printf("Creating workflow run [owner: %s, repo: %s, file: %s, ref: %s]\n", owner, repo, file, ref)

	is := make(map[string]interface{})
	for _, i := range inputs {
		is[i.Name] = i.Value
	}

	_, err := c.github.Actions.CreateWorkflowDispatchEventByFileName(context.Background(), owner, repo, file, github.CreateWorkflowDispatchEventRequest{
		Ref:    ref,
		Inputs: is,
	})
	return err
}

func (c *Client) GetWorkflowRun(owner, repo, file string, completed bool) (*github.WorkflowRun, error) {
	log.Printf("Getting workflow run [owner: %s, repo: %s, file: %s, completed: %v]\n", owner, repo, file, completed)

	opt := &github.ListWorkflowRunsOptions{
		ListOptions: github.ListOptions{PerPage: 1},
	}
	if completed {
		opt.Status = "completed"
	}
	r, _, err := c.github.Actions.ListWorkflowRunsByFileName(context.Background(), owner, repo, file, opt)
	if err != nil {
		return nil, err
	}
	if len(r.WorkflowRuns) == 0 {
		return nil, nil
	}
	return r.WorkflowRuns[0], nil
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
	log.Printf("Getting workflow run logs [owner: %s, repo: %s, id: %v]\n", owner, repo, id)

	url, _, err := c.github.Actions.GetWorkflowRunLogs(context.Background(), owner, repo, id, true)
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
	return logs, nil
}

func (c *Client) GetRelease(owner, repo, tag string) (*github.RepositoryRelease, error) {
	log.Printf("Getting release [owner: %s, repo: %s, tag: %s]\n", owner, repo, tag)

	r, _, err := c.github.Repositories.GetReleaseByTag(context.Background(), owner, repo, tag)
	if err != nil && strings.Contains(err.Error(), "404") {
		return nil, nil
	}
	return r, err
}

func (c *Client) CreateRelease(owner, repo, tag, name, body string, prerelease bool) (*github.RepositoryRelease, error) {
	log.Printf("Creating release [owner: %s, repo: %s, tag: %s]\n", owner, repo, tag)

	r, _, err := c.github.Repositories.CreateRelease(context.Background(), owner, repo, &github.RepositoryRelease{
		TagName:    &tag,
		Name:       &name,
		Body:       &body,
		Prerelease: &prerelease,
	})
	return r, err
}

func (c *Client) GetOrCreateRelease(owner, repo, tag, name, body string, prerelease bool) (*github.RepositoryRelease, error) {
	log.Printf("Getting or creating release [owner: %s, repo: %s, tag: %s]\n", owner, repo, tag)

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
	log.Printf("Getting tag [owner: %s, repo: %s, tag: %s]\n", owner, repo, tag)

	t, _, err := c.github.Git.GetTag(context.Background(), owner, repo, tag)
	if err != nil && strings.Contains(err.Error(), "404") {
		return nil, nil
	}
	return t, err
}

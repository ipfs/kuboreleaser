package git

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Client struct {
	name   string
	email  string
	auth   *HeaderAuth
	entity *openpgp.Entity
}

func NewClient() (*Client, error) {
	name := os.Getenv("GITHUB_USER_NAME")
	if name == "" {
		name = "Kubo Releaser"
	}
	email := os.Getenv("GITHUB_USER_EMAIL")
	if email == "" {
		email = "noreply+kuboreleaser@ipfs.tech"
	}
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("env var GITHUB_TOKEN must be set")
	}
	key64 := os.Getenv("GPG_KEY")
	if key64 == "" {
		return nil, fmt.Errorf("env var GPG_KEY must be set")
	}
	pass := os.Getenv("GPG_PASSPHRASE")

	// create HeaderAuth
	auth, err := NewHeaderAuth(token)
	if err != nil {
		return nil, err
	}

	// create OpenPGP Entity
	key, err := base64.StdEncoding.DecodeString(key64)
	if err != nil {
		return nil, err
	}
	bass := []byte(pass)
	list, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(key))
	if err != nil {
		return nil, err
	}
	entity := list[0]
	err = entity.PrivateKey.Decrypt(bass)
	if err != nil {
		return nil, err
	}
	for _, subkey := range entity.Subkeys {
		err = subkey.PrivateKey.Decrypt(bass)
		if err != nil {
			return nil, err
		}
	}

	return &Client{
		name:   name,
		email:  email,
		auth:   auth,
		entity: entity,
	}, nil
}

func (c *Client) signature() *object.Signature {
	return &object.Signature{
		Name:  c.name,
		Email: c.email,
		When:  time.Now(),
	}
}

type Clone struct {
	client     *Client
	repository *git.Repository
	dir        string
}

func (c *Client) Clone(dir, owner, repo, branch, sha string) (*Clone, error) {
	log.Printf("Cloning [dir: %s, owner: %s, repo: %s, branch: %s, sha: %s]\n", dir, owner, repo, branch, sha)

	log.Println("Initializing git repository")
	repository, err := git.PlainInit(dir, false)
	if err != nil {
		return nil, err
	}

	log.Println("Creating remote")
	remote, err := repository.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"https://github.com/" + owner + "/" + repo},
	})
	if err != nil {
		return nil, err
	}

	log.Println("Fetching remote")
	// https://github.com/go-git/go-git/issues/264
	err = remote.Fetch(&git.FetchOptions{
		Auth: c.auth,
		RefSpecs: []config.RefSpec{
			config.RefSpec("+" + sha + ":refs/remotes/origin/" + branch),
		},
		Tags:  git.NoTags,
		Depth: 1,
	})
	if err != nil {
		return nil, err
	}

	log.Println("Checking out branch")
	worktree, err := repository.Worktree()
	if err != nil {
		return nil, err
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Hash:   plumbing.NewHash(sha),
		Branch: plumbing.NewBranchReferenceName(branch),
		Create: true,
	})
	if err != nil {
		return nil, err
	}

	return &Clone{
		client:     c,
		repository: repository,
		dir:        dir,
	}, nil
}

func (c *Clone) Commit(glob, message string) (*object.Commit, error) {
	log.Printf("Committing [glob: %s, message: %s]\n", glob, message)

	log.Println("Adding files")
	worktree, err := c.repository.Worktree()
	if err != nil {
		return nil, err
	}
	err = worktree.AddWithOptions(&git.AddOptions{
		Glob: glob,
	})
	if err != nil {
		return nil, err
	}

	log.Println("Committing")
	hash, err := worktree.Commit(message, &git.CommitOptions{
		Author: c.client.signature(),
	})
	if err != nil {
		return nil, err
	}

	return c.repository.CommitObject(hash)
}

func (c *Clone) Tag(ref, tag, message string) (*object.Tag, error) {
	log.Printf("Tagging [ref: %s, tag: %s, message: %s]\n", ref, tag, message)

	log.Println("Creating tag")
	obj, err := c.repository.CreateTag(tag, plumbing.NewHash(ref), &git.CreateTagOptions{
		Tagger:  c.client.signature(),
		Message: message,
		SignKey: c.client.entity,
	})
	if err != nil {
		return nil, err
	}

	return c.repository.TagObject(obj.Hash())
}

func (c *Clone) Push(ref string) error {
	log.Printf("Pushing [ref: %s]\n", ref)

	log.Println("Pushing")
	return c.repository.Push(&git.PushOptions{
		Auth:       c.client.auth,
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{
			config.RefSpec(ref),
		},
	})
}

func (c *Clone) PushBranch(branch string) error {
	return c.Push(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch))
}

func (c *Clone) PushTag(tag string) error {
	return c.Push(fmt.Sprintf("refs/tags/%s:refs/tags/%s", tag, tag))
}

type Command struct {
	Name string
	Args []string
}

func (c *Command) Run(dir string) error {
	cmd := exec.Command(c.Name, c.Args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *Client) WithClone(owner, repo, branch, sha string, fn func(*Clone) error) error {
	dir, err := os.MkdirTemp("", "kuboreleaser")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	r, err := c.Clone(dir, owner, repo, branch, sha)
	if err != nil {
		return err
	}

	return fn(r)
}

func (c *Client) WithCloneCommitAndPush(owner, repo, branch, sha, glob, message string, fn func(*Clone) error) error {
	return c.WithClone(owner, repo, branch, sha, func(r *Clone) error {
		err := fn(r)
		if err != nil {
			return err
		}

		_, err = r.Commit(glob, message)
		if err != nil {
			return err
		}

		return r.PushBranch(branch)
	})
}

func (c *Client) WithCloneExecCommitAndPush(owner, repo, branch, sha, glob, message string, commands ...Command) error {
	return c.WithCloneCommitAndPush(owner, repo, branch, sha, glob, message, func(r *Clone) error {
		for _, command := range commands {
			err := command.Run(r.dir)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

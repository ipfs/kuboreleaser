package git

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/ipfs/kuboreleaser/util"
	log "github.com/sirupsen/logrus"

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
		return nil, fmt.Errorf("GITHUB_TOKEN not set")
	}

	// create HeaderAuth
	auth, err := NewHeaderAuth(token)
	if err != nil {
		return nil, err
	}

	disabled := os.Getenv("NO_GPG")
	if disabled != "" {
		return &Client{
			name:  name,
			email: email,
			auth:  auth,
		}, nil
	}

	key64 := os.Getenv("GPG_KEY")
	if key64 == "" {
		return nil, fmt.Errorf("GPG_KEY not set")
	}
	pass := os.Getenv("GPG_PASSPHRASE")

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
	log.WithFields(log.Fields{
		"dir":    dir,
		"owner":  owner,
		"repo":   repo,
		"branch": branch,
		"sha":    sha,
	}).Debug("Cloning...")

	log.Debug("Initializing git repository...")
	repository, err := git.PlainInit(dir, false)
	if err != nil {
		return nil, err
	}

	log.Debug("Adding remote...")
	remote, err := repository.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"https://github.com/" + owner + "/" + repo},
	})
	if err != nil {
		return nil, err
	}

	log.Debug("Fetching...")
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

	log.Debug("Checking out...")
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

	log.Debug("Cloned")

	return &Clone{
		client:     c,
		repository: repository,
		dir:        dir,
	}, nil
}

func (c *Clone) Status() (git.Status, error) {
	log.Debug("Retrieving status...")

	worktree, err := c.repository.Worktree()
	if err != nil {
		return nil, err
	}
	status, err := worktree.Status()
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"status": status,
	}).Debug("Retrieved status")

	return status, nil
}

func (c *Clone) Commit(glob, message string) (*object.Commit, error) {
	log.WithFields(log.Fields{
		"glob":    glob,
		"message": message,
	}).Debug("Committing...")

	log.Debug("Adding files...")
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

	log.Debug("Creating commit...")
	hash, err := worktree.Commit(message, &git.CommitOptions{
		Author: c.client.signature(),
	})
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"hash": hash,
	}).Debug("Commit created")

	return c.repository.CommitObject(hash)
}

func (c *Clone) Tag(ref, tag, message string) (*object.Tag, error) {
	log.WithFields(log.Fields{
		"ref": ref,
		"tag": tag,
	}).Debug("Tagging...")

	options := &git.CreateTagOptions{
		Tagger:  c.client.signature(),
		Message: message,
	}
	if c.client.entity != nil {
		options.SignKey = c.client.entity
	} else {
		log.Warn("No OpenPGP key found, tag will not be signed")
	}

	log.Debug("Creating tag...")
	obj, err := c.repository.CreateTag(tag, plumbing.NewHash(ref), options)
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"hash": obj.Hash(),
	}).Debug("Tag created")

	return c.repository.TagObject(obj.Hash())
}

func (c *Clone) Push(ref string) error {
	log.WithFields(log.Fields{
		"ref": ref,
	}).Debug("Pushing...")

	err := c.repository.Push(&git.PushOptions{
		Auth:       c.client.auth,
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{
			config.RefSpec(ref),
		},
	})

	if err != nil {
		log.Debug("Pushed")
	}

	return err
}

func (c *Clone) PushBranch(branch string) error {
	return c.Push(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch))
}

func (c *Clone) PushTag(tag string) error {
	return c.Push(fmt.Sprintf("refs/tags/%s:refs/tags/%s", tag, tag))
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

func (c *Client) RunAndPush(owner, repo, branch, sha, message string, commands ...util.Command) error {
	return c.WithClone(owner, repo, branch, sha, func(r *Clone) error {
		for _, command := range commands {
			command.Dir = r.dir
			err := command.Run()
			if err != nil {
				return err
			}
		}

		status, err := r.Status()
		if err != nil {
			return err
		}

		if !status.IsClean() {
			_, err = r.Commit("*", message)
			if err != nil {
				return err
			}

			return r.PushBranch(branch)
		}

		return nil
	})
}

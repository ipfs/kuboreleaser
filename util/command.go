package util

import (
	"io"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type Stdout struct {
	Writer io.Writer
}

func (s Stdout) Write(p []byte) (n int, err error) {
	os.Stdout.Write(p)
	return s.Writer.Write(p)
}

type Command struct {
	Name   string
	Args   []string
	Dir    string
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
	Env    []string
}

func (c *Command) Run() error {
	log.WithFields(log.Fields{
		"name": c.Name,
		"args": c.Args,
		"dir":  c.Dir,
	}).Debug("Running command...")

	cmd := exec.Command(c.Name, c.Args...)
	if c.Dir != "" {
		cmd.Dir = c.Dir
	}
	if c.Stdout != nil {
		cmd.Stdout = c.Stdout
	} else {
		cmd.Stdout = os.Stdout
	}
	if c.Stderr != nil {
		cmd.Stderr = c.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}
	if c.Stdin != nil {
		cmd.Stdin = c.Stdin
	}
	if c.Env != nil {
		cmd.Env = c.Env
	}
	return cmd.Run()
}

package repos

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/util"
)

type ipfsDocs struct {
	Owner           string
	Repo            string
	DefaultBranch   string
	WorkflowName    string
	WorkflowJobName string
}

var IPFSDocs = ipfsDocs{
	Owner:           "ipfs",
	Repo:            "ipfs-docs",
	DefaultBranch:   "main",
	WorkflowName:    "update-on-new-ipfs-tag.yml",
	WorkflowJobName: "update",
}

func (i ipfsDocs) KuboBranch(version *util.Version) string {
	return fmt.Sprintf("kubo-%s", version)
}

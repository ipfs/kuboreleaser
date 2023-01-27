package repos

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/util"
)

type ipfsBlog struct {
	Owner         string
	Repo          string
	DefaultBranch string
}

var IPFSBlog = ipfsBlog{
	Owner:         "ipfs",
	Repo:          "ipfs-blog",
	DefaultBranch: "main",
}

func (i ipfsBlog) KuboBranch(version *util.Version) string {
	return fmt.Sprintf("kubo-%s", version)
}

package repos

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/util"
)

type interop struct {
	Owner         string
	Repo          string
	DefaultBranch string
}

var Interop = interop{
	Owner:         "ipfs",
	Repo:          "interop",
	DefaultBranch: "master",
}

func (i interop) KuboBranch(version *util.Version) string {
	return fmt.Sprintf("kubo-%s", version)
}

package repos

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/util"
)

type distributions struct {
	Owner         string
	Repo          string
	DefaultBranch string
}

var Distributions = distributions{
	Owner:         "ipfs",
	Repo:          "distributions",
	DefaultBranch: "master",
}

func (d distributions) KuboBranch(version *util.Version) string {
	return fmt.Sprintf("kubo-%s", version)
}

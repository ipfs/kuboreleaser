package repos

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/util"
)

type ipfsDesktop struct {
	Owner         string
	Repo          string
	DefaultBranch string
}

var IPFSDesktop = ipfsDesktop{
	Owner:         "ipfs",
	Repo:          "ipfs-desktop",
	DefaultBranch: "main",
}

func (i ipfsDesktop) KuboBranch(version *util.Version) string {
	return fmt.Sprintf("kubo-%s", version)
}

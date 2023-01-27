package repos

type ipfsCompanion struct {
	Owner           string
	Repo            string
	DefaultBranch   string
	WorkflowName    string
	WorkflowJobName string
}

var IPFSCompanion = ipfsCompanion{
	Owner:           "ipfs",
	Repo:            "ipfs-companion",
	DefaultBranch:   "main",
	WorkflowName:    "e2e.yml",
	WorkflowJobName: "test",
}

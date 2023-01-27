package repos

type npmGoIPFS struct {
	Owner           string
	Repo            string
	DefaultBranch   string
	WorkflowName    string
	WorkflowJobName string
}

var NPMGoIPFS = npmGoIPFS{
	Owner:           "ipfs",
	Repo:            "npm-go-ipfs",
	DefaultBranch:   "master",
	WorkflowName:    "main.yml",
	WorkflowJobName: "publish",
}

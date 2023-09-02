package repos

type npmKubo struct {
	Owner           string
	Repo            string
	DefaultBranch   string
	WorkflowName    string
	WorkflowJobName string
}

var NPMKubo = npmKubo{
	Owner:           "ipfs",
	Repo:            "npm-kubo",
	DefaultBranch:   "master",
	WorkflowName:    "main.yml",
	WorkflowJobName: "publish",
}

package repos

type boxo struct {
	Owner                            string
	Repo                             string
	DefaultBranch                    string
}

var Boxo = boxo{
	Owner:                            "ipfs",
	Repo:                             "boxo",
	DefaultBranch:                    "main",
}

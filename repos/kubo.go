package repos

import (
	"fmt"

	"github.com/ipfs/kuboreleaser/util"
)

type kubo struct {
	Owner                            string
	Repo                             string
	DefaultBranch                    string
	ReleaseBranch                    string
	SyncReleaseAssetsWorkflowName    string
	SyncReleaseAssetsWorkflowJobName string
}

var Kubo = kubo{
	Owner:                            "ipfs",
	Repo:                             "kubo",
	DefaultBranch:                    "master",
	ReleaseBranch:                    "release",
	SyncReleaseAssetsWorkflowName:    "sync-release-assets.yml",
	SyncReleaseAssetsWorkflowJobName: "sync-github-and-dist-ipfs-tech",
}

func (k kubo) VersionReleaseBranch(version *util.Version) string {
	return fmt.Sprintf("release-%s", version.MajorMinorPatch())
}

func (k kubo) VersionUpdateBranch(version *util.Version) string {
	return fmt.Sprintf("version-update-%s", version.MajorMinor())
}

func (k kubo) ReleaseMergeBranch(version *util.Version) string {
	return fmt.Sprintf("merge-release-%s", version.MajorMinorPatch())
}

func (k kubo) ReleaseIssueTitle(version *util.Version) string {
	return fmt.Sprintf("Release %s", version.MajorMinorPatch()[1:])
}

package repos

import (
	"fmt"
	"strings"

	"github.com/ipfs/kuboreleaser/util"
)

type kubo struct {
	Owner                            string
	Repo                             string
	DefaultBranch                    string
	ReleaseBranch                    string
	SyncReleaseAssetsWorkflowName    string
	SyncReleaseAssetsWorkflowJobName string
	DockerHubWorkflowName            string
	DockerHubWorkflowJobName         string
}

var Kubo = kubo{
	Owner:                            "ipfs",
	Repo:                             "kubo",
	DefaultBranch:                    "master",
	ReleaseBranch:                    "release",
	SyncReleaseAssetsWorkflowName:    "sync-release-assets.yml",
	SyncReleaseAssetsWorkflowJobName: "dist-ipfs-tech",
	DockerHubWorkflowName:            "docker-image.yml",
	DockerHubWorkflowJobName:         "Push Docker image to Docker Hub",
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

func (k kubo) ChangelogBranch(version *util.Version) string {
	return fmt.Sprintf("changelog-%s", version.MajorMinorPatch())
}

func (k kubo) ReleaseIssueTitle(version *util.Version) string {
	return fmt.Sprintf("Release %s", strings.TrimSuffix(version.MajorMinorPatch()[1:], ".0"))
}

func (k kubo) ReleaseURL(version *util.Version) string {
	return fmt.Sprintf("https://github.com/%s/%s/releases/tag/%s", k.Owner, k.Repo, version)
}

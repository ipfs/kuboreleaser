# KuboReleaser

KuboReleaser is a CLI tool intended to help with the automation of [the release process of Kubo](https://github.com/ipfs/kubo/blob/master/docs/RELEASE_ISSUE_TEMPLATE.md).

It was originally started here - https://github.com/ipfs/kubo/pull/9493 - and is now being extracted into its own repo.

## TODO

- [ ] enable auto-merge on created PRs
- [ ] assign reviewers to the created PRs
- [ ] check how git-go performs fetch (does it use protocol.version 2?)
- [ ] handle the case where PR was closed without being merged
- [ ] add a `--dry-run` flag
- [ ] change PR to ready for review on final release (e.g. in cut-branch)
- [ ] automate changelog update (mkreleaselog; might require git in docker)
- [ ] retrieving check runs fails if the branch was deleted
- [ ] link to the release in bifrost comms
- [ ] link to the release issue in the PRs
- [ ] do not check runs status if the PR is merged

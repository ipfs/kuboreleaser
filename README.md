# KuboReleaser

KuboReleaser is a CLI tool intended to help with the automation of [the release process of Kubo](https://github.com/ipfs/kubo/blob/master/docs/RELEASE_ISSUE_TEMPLATE.md).

It was originally started here - https://github.com/ipfs/kubo/pull/9493 - and is now being extracted into its own repo.

## TODO

- [ ] imlement exponential backoff to recover from `CheckErrorWait` errors
- [ ] add `--skip-wait` flag
- [ ] test the CLI App
- [ ] enable auto-merge on created PRs
- [ ] assign reviewers to the created PRs
- [ ] check how git-go performs fetch (does it use protocol.version 2?)
- [ ] run exec commands with streaming output
- [ ] handle the case where PR was closed without being merged
- [ ] add a `--dry-run` flag
- [ ] add wait for CI checks to appear (e.g. in cut-branch it seems that automation is too fast)
- [ ] change PR to ready for review on final release (e.g. in cut-branch)
- [ ] automate changelog creation (last step of previous release)
- [ ] automate changelog update (mkreleaselog; might require git in docker)
- [ ] wait after tag push (seems that github is not as fast as git)
- [ ] add check for docker image publishing (somewhere?)
- [ ] account for checks that are skipped (e.g diff in publish-to-distributions@master)
- [ ] handle patch releases in cut branch
- [ ] tell user to merge PRs, e.g. in publish to dists

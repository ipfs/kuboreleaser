# KuboReleaser

KuboReleaser is a CLI tool intended to help with the automation of [the release process of Kubo](https://github.com/ipfs/kubo/blob/master/docs/RELEASE_ISSUE_TEMPLATE.md).

It was originally started here - https://github.com/ipfs/kubo/pull/9493 - and is now being extracted into its own repo.

It is not yet usable since none of the commands are hooked up to the CLI App.

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

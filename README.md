# KuboReleaser

KuboReleaser is a CLI tool intended to help with the automation of [the release process of Kubo](https://github.com/ipfs/kubo/blob/master/docs/RELEASE_ISSUE_TEMPLATE.md).

It was originally started here - https://github.com/ipfs/kubo/pull/9493 - and is now being extracted into its own repo.

## TODO

- [ ] enable auto-merge on created PRs
- [ ] assign reviewers to the created PRs
- [ ] check how git-go performs fetch (does it use protocol.version 2?)
- [ ] add a `--dry-run` flag
- [ ] automate changelog update (mkreleaselog; might require git in docker)
- [ ] link to the release issue in the PRs
- [ ] allow to specify args via env vars
- [ ] remove one level of nesting from the CLI
- [ ] document how to use kuboreleaser in the README

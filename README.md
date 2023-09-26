# KuboReleaser

KuboReleaser is a CLI tool intended to help with the automation of [the release process of Kubo](https://github.com/ipfs/kubo/blob/master/docs/RELEASE_ISSUE_TEMPLATE.md).

It was originally started here - https://github.com/ipfs/kubo/pull/9493 - and is now being extracted into its own repo.

## Prerequisites

- [ ] `docker` installed
- [ ] GitHub token creted with the following scopes
  - ...
- [ ] GitHub GPG key created and added to the GitHub account
- [ ] Matrix account created and added to the Kubo room

## Usage

1. Build the binary

```bash
make kuboreleaser
```

2. Create the .env file with your credentials

```bash
make env
```

3. Run the CLI

```bash
./kubeleaser --help
```

## Other

You can skip GPG setup by exporting `NO_GPG=true` in your environment. If you do that, you won't be able to sign the release tag.

You can skip Matrix setup by exporting `NO_MATRIX=true` in your environment. If you do that, you will have to confirm promotional posts were posted to Matrix manually.

## TODO

- [ ] enable auto-merge on created PRs
- [ ] assign reviewers to the created PRs
- [ ] check how git-go performs fetch (does it use protocol.version 2?)
- [ ] add a `--dry-run` flag
- [ ] link to the release issue in the PRs
- [ ] allow to specify args via env vars
- [ ] remove one level of nesting from the CLI
- [ ] document how to use kuboreleaser in the README

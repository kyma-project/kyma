# Changelog Generator

## Overview

This project is a Docker image that is used to generate a changelog in the `kyma` repository. It uses GitHub API to get pull requests with specified labels.

## Prerequisites

To set up the project, you need the latest version of [Docker](https://www.docker.com/).

## Usage

Read how to build the Docker image and use it to generate either the full changelog with all released Kyma versions or a changelog only for the latest release.  

### Build the Docker image

To build the Docker image, run this command:

```bash
docker build -t changelog-generator .
```

### Generate the full changelog

To generate a `CHANGELOG.md` file for all released Kyma versions, follow these steps:

1. Create a `CHANGELOG.md` file under the specified absolute path to the repository. Use this command:

```bash
docker run --rm -v {absolutePathToRepository}:/repository -w /repository -e LATEST_VERSION={applicationVersion} -e GITHUB_AUTH={githubToken} -e SSH_FILE={sshFilePath} -e CONFIG_FILE={configFilePath} -e SKIP_REMOVING_LATEST=true changelog-generator sh /app/generate-full-changelog.sh --configure-git
```

Replace values in curly braces with proper details, where:
- `{absolutePathToRepository}` is the absolute path to the repository.
- `{applicationVersion}` is the currently released application version.
- `{githubToken}` is the GitHub API token with the read-only access to the repository.
- `{sshFilePath}` is the path to the SSH file used for Git to authenticate with the repository.
- `{configFilePath}` is the path to file used by `lerna-changelog` to repository configuration

Add `SKIP_REMOVING_LATEST=true` variable if you want skip removing 'latest' tag functionality.

2. Commit and push the `CHANGELOG.md` file. Use this command:

```bash
docker run --rm -v {absolutePathToRepository}:/repository -w /repository -e BRANCH={currentBranch} -e LATEST_VERSION={applicationVersion} -e SSH_FILE={sshFilePath} changelog-generator sh /app/push-full-changelog.sh --configure-git
```

Replace values in curly braces with proper details, where:
- `{absolutePathToRepository}` is the absolute path to the repository.
- `{currentBranch}` is the current Git branch name.
- `{applicationVersion}` is the currently released application version.
- `{sshFilePath}` is the path to the SSH file used for Git to authenticate with the repository.

### Generate the latest release changelog

To generate a changelog for a single release that contains merged pull requests for the latest Kyma version, run this command:

```bash
docker run --rm -v /path/to/repository/:/repository -w /repository -e FROM_TAG={previousTag} -e LATEST_VERSION={applicationVersion} -e GITHUB_AUTH={githubToken} -e SSH_FILE={sshFile} -e CONFIG_FILE={configFilePath} changelog-generator sh /app/generate-release-changelog.sh --configure-git
```

Replace values in curly braces with proper details, where:
- `{absolutePathToRepository}` is the absolute path to the repository.
- `{previousTag}` optionally, one tag before last. If provided changelog will be generated from `{previousTag}` to `{applicationVersion}`
- `{applicationVersion}` is the currently released application version.
- `{githubToken}` is the GitHub API token with the read-only access to the repository.
- `{sshFilePath}` is the path to the SSH file used for Git to authenticate with the repository.
- `{configFilePath}` is the path to file used by `lerna-changelog` to repository configuration

Add `SKIP_REMOVING_LATEST=true` variable if you want skip removing 'latest' tag functionality.

The script generates a new `./.changelog/release-changelog.md` file under the specified absolute path to the repository.

## Development

This project uses the [`lerna-changelog`](https://github.com/lerna/lerna-changelog) generator and custom shell scripts to create a release changelog in the form of a `CHANGELOG.md` file.

### Configure changelog generation

To specify the changelog generation settings, modify the `package.json` file under the `app` directory. See the [`lerna-changelog` documentation](https://github.com/lerna/lerna-changelog/blob/master/README.md) for details.
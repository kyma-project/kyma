---
title: Git source type
---

Depending on a runtime you use to build your Function (Node.js 12, Node.js 14, or Python 3.8), your Git repository must contain at least a directory with these files:

- `handler.js` or `handler.py` with Function's code
- `package.json` or `requirements.txt` with Function's dependencies

The Function CR must contain `type: git` to specify that you use a Git repository for the Function's sources.

To create a Function with the Git source, you must:

1. Create a [GitRepository CR](./05-technical-reference/06-custom-resources/svls-02-gitrepository.md) with details of your Git repository.
2. Create a [Secret](https://kubernetes.io/docs/concepts/configuration/secret/) (optional, only if you must authenticate to the repository).
3. Create a [Function CR](./05-technical-reference/06-custom-resources/svls-01-function.md) with your Function definition and references to the Git repository.

>**NOTE:** For detailed steps, see the tutorial on [creating a Function from Git repository sources](./03-tutorials/serverless/svls-02-create-git-function.md).

You can have various setups for your Function's Git source with different:

- Directory structures

  You can specify the location of your code dependencies with the **baseDir** parameter in the Function CR. For example, use `"/"` if you keep the source files at the root of your repository.

- Authentication methods

  You can define with the **spec.auth** parameter in the GitRepository CR that you must authenticate to the repository with a password or token (`basic`), or an SSH key (`key`).

- Function's rebuild triggers

  You can use the **reference** parameter in the GitRepository CR to define whether the Function Controller must monitor a given branch or commit in the Git repository to rebuild the Function upon their changes.

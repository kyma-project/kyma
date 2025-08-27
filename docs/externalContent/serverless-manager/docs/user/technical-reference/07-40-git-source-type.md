# Git Source Type

Depending on a runtime you use to build your Function (Node.js or Python), your Git repository must contain at least a directory with these files:

- `handler.js` or `handler.py` with Function's code
- `package.json` or `requirements.txt` with Function's dependencies

The Function CR must contain **spec.source.gitRepository** to specify that you use a Git repository for the Function's sources.

To create a Function with the Git source, you must:

1. Create a [Secret](https://kubernetes.io/docs/concepts/configuration/secret/) (optional, only if you must authenticate to the repository).
2. Create a [Function CR](../resources/06-10-function-cr.md) with your Function definition and references to the Git repository.

> [!NOTE]
> For detailed steps, see the tutorial on [creating a Function from Git repository sources](../tutorials/01-11-create-git-function.md).

You can have various setups for your Function's Git source with different:

- Directory structures

  To specify the location of your code dependencies, use the **baseDir** parameter in the Function CR. For example, use `"/"` if you keep the source files at the root of your repository.

- Authentication methods

  To define that you must authenticate to the repository with a password or token (`basic`), or an SSH key (`key`), use the **spec.source.gitRepository.auth** parameter in the Function CR.

- Function's rebuild triggers

  To define whether the Function Controller must monitor a given branch or commit in the Git repository to rebuild the Function upon their changes, use the **spec.source.gitRepository.reference** parameter in the Function CR.
  
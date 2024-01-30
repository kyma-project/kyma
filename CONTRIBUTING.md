## Overview

To contribute to this project, follow the rules from the general [CONTRIBUTING.md](https://github.com/kyma-project/community/blob/main/CONTRIBUTING.md) document in the `community` repository.

## Contribution Rules

Apart from the general rules described in the `community` repository, every `kyma` repository contributor must follow these basic rules:

* Do not copy charts from the Internet. Customize the Helm charts and simplify them to pertain only to the specific use case. Apply this rule to all documents associated with the charts, such as `README.md` and `NOTES.txt` documents.
* Follow the `IfNotPresent` pulling policy. Do not use the `latest` tag for all `Deployments` definitions for the local installation.
* Adjust any data copied from the Internet to the product needs.
* When you receive the required approvals for your pull request and it is merged, create another one to update the image version and any configuration changes in relevant Kyma charts. The Continuous Integration system generates and pushes images. You must implement the configuration changes manually.

## Development

1. Develop on your remote repository forked from the original repository in Go.

   Follow these steps:


   > [!NOTE]
   > The example assumes you have the `$GOPATH` already set.

1. Fork the repository in GitHub.

2. Clone the fork to your `$GOPATH` workspace. Use this command to create the folder structure and clone the repository under the correct location:

    ```bash
    git clone git@github.com:{GitHubUsername}/kyma.git $GOPATH/src/github.com/kyma-project/kyma
    ```

    Follow the steps described in the [`git-workflow.md`](https://github.com/kyma-project/community/blob/main/docs/contributing/03-git-workflow.md) document to configure your fork.

3. Build the project.

    Every project runs differently. Follow instructions in the main `README.md` document of the given project to build it.

4. Create a branch and start to develop.

    Do not forget about creating unit and acceptance tests if needed. For the unit tests, follow the instructions specified in the main `README.md` document of the given project. For the details concerning the acceptance tests, go to the corresponding directory inside the `tests` directory.

5. Test your changes.
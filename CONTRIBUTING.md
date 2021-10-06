## Overview

To contribute to this project, follow the rules from the general [CONTRIBUTING.md](https://github.com/kyma-project/community/blob/main/CONTRIBUTING.md) document in the `community` repository.

## Contribution rules

Apart from the general rules described in the `community` repository, every `kyma` repository contributor must follow these basic rules:

* Do not copy charts from the Internet. Customize the Helm charts and simplify them to pertain only to the specific use case. Apply this rule to all documents associated with the charts, such as `README.md` and `NOTES.txt` documents.
* Follow the `IfNotPresent` pulling policy. Do not use the `latest` tag for all `Deployments` definitions for the local installation.
* Adjust any data copied from the Internet to the product needs.
* When you receive the required approvals for your pull request and it is merged, create another one to update the image version and any configuration changes in relevant Kyma charts. The Continuous Integration system generates and pushes images. You must implement the configuration changes manually.

## Overview

To contribute to this project, follow the rules from the general [CONTRIBUTING.md](https://github.com/kyma-project/community/blob/master/CONTRIBUTING.md) document in the `community` repository.

## Documentation types

These are the main types of documents used in the project:

* `NOTES.txt`- This document type is an integral part of Helm charts. Its content displays in the terminal window after installing the chart. It is not mandatory to include `NOTES.txt` documents for sub-charts because the system ignores these documents. Provide `NOTES.txt` documents for the Core components. Use the [template](https://github.com/kyma-project/community/blob/master/guidelines/templates/resources/NOTES.txt) to create `NOTES.txt` documents.

* `README.md` - This document type contains information about other files in the directory. Each main directory in this repository, such as `cluster` or `resources`, requires a `README.md` document. Additionally, each chart and sub-chart needs such a document. Add a `README.md` document when you create a new directory or chart. Use the [template](https://github.com/kyma-project/community//blob/master/guidelines/templates/resources/chart_README.md) to create `README.md` documents.

Do not change the names or the order of the main sections in the `README.md` documents. However, you can create subsections to adjust each `README.md` document to the project's or chart's specific requirements. See the example of a [README.md](resources/core/README.md) document.

## Contribution rules

Apart from the general rules described in the `community` repository, every `kyma` repository contributor must follow these basic rules:

* Do not copy charts from the Internet. Customize the Helm charts and simplify them to pertain only to the specific use case. Apply this rule to all documents associated with the charts, such as `README.md` and `NOTES.txt` documents.
* Follow the `IfNotPresent` pulling policy. Do not use the `latest` tag for all `Deployments` definitions for the local installation.
* Adjust any data copied from the Internet to the product needs.
* When you receive the required approvals for your pull request and it is merged, create another one to update the image version and any configuration changes in relevant Kyma charts. The Continuous Integration system generates and pushes images. You must implement the configuration changes manually.

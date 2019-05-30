a---
title: Bundles validation
type: Details
---

The Helm Broker validates the bundles which fetches. If some bundle does not meet the requirements, the Helm Broker won't expose it as a Service Class and it will put an information about this in the logs.

To create a correct bundle follow the following rules:

1. The bundle **must** contains a `meta.yaml` file with non-empty values in the following fields:

- id
- name
- version (compatible with [Semantic Versioning](https://semver.org/))
- description
- displayName

2. Bundle **must** contain at least one plan. Each plan **must** follow the rules:

- If plan `bindable` value in the `meta.yaml` file is set to `true`, the plan **must** contains the `bind.yaml` file which defines the data of the Service Bindings.

The `meta.yaml` file of the plan **must** contain non-empty values in the following fields:

- id
- name
- description
- displayName

3. If the bundle provides documentation for its Service Class it **must** contains the `meta.yaml` file in the `docs` directory. The `meta.yaml` file **must** contain a **single** entry in the `docs` array.

## Checker

The Checker tool is used in the [bundles](https://github.com/kyma-project/bundles) repo. It is triggered there for each pull request. It triggers the validation logic described above for all of the bundles. It's also triggering the `helm lint` command which check the bundle's chart.
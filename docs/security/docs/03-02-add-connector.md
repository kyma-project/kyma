---
title: Add an Identity Provider to Dex
type: Details
---

Add external, OpenID Connect compliant authentication providers to Kyma using [Dex connectors](https://github.com/coreos/dex#connectors). Follow the instructions below to add a GitHub connector and use it to authenticate users in Kyma.

>**NOTE:** Groups in the Github are represented as teams. See [this](https://help.github.com/articles/organizing-members-into-teams/) document to learn how to manage teams in Github.

## Prerequisites

To add a GitHub connector to Dex, [register](https://github.com/settings/applications/new) a new OAuth application in GitHub. Set the authorization callback URL to `https://dex.kyma.local/callback`.
After you complete the registration, [request](https://help.github.com/articles/requesting-organization-approval-for-oauth-apps/) for an organization approval.

>**NOTE:** To authenticate in Kyma using GitHub, the user must be a member of a GitHub [organization](https://help.github.com/articles/creating-a-new-organization-from-scratch/) that has at least one [team](https://help.github.com/articles/creating-a-team/).

## Configure Dex

Register the GitHub Dex connector by editing the `dex-config-map.yaml` ConfigMap file located in the `kyma/resources/dex/templates` directory. Follow this template:

```
    connectors:
    - type: github
      id: github
      name: GitHub
      config:
        clientID: {GITHUB_CLIENT_ID}
        clientSecret: {GITHUB_CLIENT_SECRET}
        redirectURI: https://dex.kyma.local/callback
        orgs:
          - name: {GITHUB_ORGANIZATION}
```

This table explains the placeholders used in the template:

|Placeholder | Description |
|---|---|
| GITHUB_CLIENT_ID | Specifies the application's client ID. |
| GITHUB_CLIENT_SECRET | Specifies the application's client Secret. |
| GITHUB_ORGANIZATION | Specifies the name of the GitHub organization. |

## Configure authorization rules

To bind Github groups to the default roles added to every Kyma Namespace, add the **bindings** section to [this](https://github.com/kyma-project/kyma/blob/master/resources/core/charts/cluster-users/values.yaml) file. Follow this template:

```
bindings:
  kymaAdmin:
    groups:
    - "{GITHUB_ORGANIZATION}:{GITHUB_TEAM_A}"
  kymaView:
    groups:
    - "{GITHUB_ORGANIZATION}:{GITHUB_TEAM_B}"
```

This table explains the placeholders used in the template:

|Placeholder | Description |
|---|---|
| GITHUB_ORGANIZATION | Specifies the name of the GitHub organization. |
| GITHUB_TEAM_A | Specifies the name of GitHub team to bind to the `kyma-admin-role` role. |
| GITHUB_TEAM_B | Specifies the name of GitHub team to bind to the `kyma-reader-role` role. |

---
title: Add an Identity Provider to Dex
type: Tutorials
---

Add external identity providers to Kyma using [Dex connectors](https://github.com/dexidp/dex#connectors). You can add connectors to Dex by creating component overrides.  
This tutorial shows how to add a [GitHub connector](https://github.com/dexidp/dex/blob/master/Documentation/connectors/github.md) and use it to authenticate users in Kyma.

>**NOTE:** Groups in the Github are represented as teams. See [this](https://help.github.com/articles/organizing-members-into-teams/) document to learn how to manage teams in Github.

## Prerequisites

To add a GitHub connector to Dex, [register](https://github.com/settings/applications/new) a new OAuth application in GitHub. Set the authorization callback URL to `https://dex.{CLUSTER_DOMAIN}/callback`.
After you complete the registration, [request](https://help.github.com/articles/requesting-organization-approval-for-oauth-apps/) for an organization approval.

>**NOTE:** To authenticate in Kyma using GitHub, the user must be a member of a GitHub [organization](https://help.github.com/articles/creating-a-new-organization-from-scratch/) that has at least one [team](https://help.github.com/articles/creating-a-team/).

## Configure Dex

Register the connector by creating a [Helm override](/docs/root/#configuration-helm-overrides-for-kyma-installation) for Dex. Create the override ConfigMap in the Kubernetes cluster before Dex is installed. If you want to register a connector at runtime, trigger the [update process](/docs/root/#installation-update-kyma-trigger-the-update-process) after creating the override.
>**TIP:** You can use Go Template expressions in the override value. These expressions are resolved by Helm using the same set of overrides as configured for the entire chart.

```
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: dex-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: dex
    kyma-project.io/installation: ""
data:
    connectors: |-
    - type: github
      id: github
      name: GitHub
      config:
        clientID: {GITHUB_CLIENT_ID}
        clientSecret: {GITHUB_CLIENT_SECRET}
         redirectURI: https://dex.{{ .Values.global.domainName }}/callback
        orgs:
          - name: {GITHUB_ORGANIZATION}
EOF
```

This table explains the placeholders used in the template:

|Placeholder | Description |
|---|---|
| GITHUB_CLIENT_ID | Specifies the application's client ID. |
| GITHUB_CLIENT_SECRET | Specifies the application's client Secret. |
| GITHUB_ORGANIZATION | Specifies the name of the GitHub organization. |

## Configure authorization rules

To bind Github groups to default Kyma roles, edit the **bindings** section in [this](https://github.com/kyma-project/kyma/blob/master/resources/core/charts/cluster-users/values.yaml) file. Follow this template:

```
bindings:
  kymaAdmin:
    groups:
    - "{GITHUB_ORGANIZATION}:{GITHUB_TEAM_A}"
  kymaView:
    groups:
    - "{GITHUB_ORGANIZATION}:{GITHUB_TEAM_B}"
```

>**TIP:** You can bind GitHub teams to any of the five predefined Kyma roles. Use these aliases: `kymaAdmin`, `kymaView`, `kymaDeveloper`, `kymaEdit` or `kymaEssentials`. To learn more about the predefined roles, read [this](#details-roles-in-kyma) document.

This table explains the placeholders used in the template:

|Placeholder | Description |
|---|---|
| GITHUB_ORGANIZATION | Specifies the name of the GitHub organization. |
| GITHUB_TEAM_A | Specifies the name of GitHub team to bind to the `kyma-admin` role. |
| GITHUB_TEAM_B | Specifies the name of GitHub team to bind to the `kyma-view` role. |

# GitHub based authentication

## Overview

In Kyma authentication providers can be easily added by using [dex connectors](https://github.com/coreos/dex#connectors). This document describes how to setup authentication throughout GitHub.

## Dex GitHub connector

### Prerequisites

It is assumed that user is a member of a GitHub [organization](https://help.github.com/articles/creating-a-new-organization-from-scratch/) with at least one [team](https://help.github.com/articles/creating-a-team/).

### Register an application

[Register](https://github.com/settings/applications/new) a new OAuth application with GitHub ensuring the authorization callback URL is `https://dex.kyma.local/callback`. Once registered [request](https://help.github.com/articles/requesting-organization-approval-for-oauth-apps/) organization approval.

### Configure dex 

Register github dex connector in `resources/core/charts/dex/templates/pre-install-dex-config-map.yaml`

```yaml
    connectors: 
    - type: github
      id: github
      name: GitHub
      config:
        clientID: ${GITHUB_CLIENT_ID}
        clientSecret: ${GITHUB_CLIENT_SECRET}
        redirectURI: https://dex.kyma.local/callback
        orgs:
          - name: ${GITHUB_ORGANIZATION}
```

where

|   |   |
|---|---|
| GITHUB_CLIENT_ID | Specifies the application client ID |
| GITHUB_CLIENT_SECRET | Specifies the application client secret |
| GITHUB_ORGANIZATION | Specifies the name of GitHub organization |

### Configure authorization rules

Authorization in Kyma is organized on the basis of [Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/). By default Kyma comes with following ClusterRoles:

- kyma-admin (full access)
- kyma-view (read-only)

To bind to those roles add following section in `resources/core/charts/cluster-users/values.yaml` 

```yaml
bindings:
  kymaAdmin:
    groups:
    - "${GITHUB_ORGANIZATION}:${GITHUB_TEAM_A}"
  kymaView:
    groups:
    - "${GITHUB_ORGANIZATION}:${GITHUB_TEAM_B}"
```

where

|   |   |
|---|---|
| GITHUB_ORGANIZATION | Specifies the name of GitHub organization |
| GITHUB_TEAM_A | Specifies the name of GitHub team that is going to be binded to kyma-admin role |
| GITHUB_TEAM_B | Specifies the name of GitHub team that is going to be binded to kyma-view role |

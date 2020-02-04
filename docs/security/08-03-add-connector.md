---
title: Add an Identity Provider to Dex
type: Tutorials
---

Add external identity providers to Kyma using [Dex connectors](https://github.com/dexidp/dex#connectors). You can add connectors to Dex by creating component overrides.
This tutorial shows how to add a [GitHub](https://github.com/dexidp/dex/blob/master/Documentation/connectors/github.md) or [XSUAA](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/ea0281368f11472b8d2b145a2a28666c.html) connector and use it to authenticate users in Kyma.

>**NOTE:** Groups in Github are represented as teams. See [this](https://help.github.com/articles/organizing-members-into-teams/) document to learn how to manage teams in Github.

## Prerequisites

<div tabs>
  <details>
  <summary>
  GitHub
  </summary>

  To add a GitHub connector to Dex, [register](https://github.com/settings/applications/new) a new OAuth2 application in GitHub. Set the authorization callback URL to `https://dex.{CLUSTER_DOMAIN}/callback`.
  After you complete the registration, [request](https://help.github.com/articles/requesting-organization-approval-for-oauth-apps/) for an organization approval.

  >**NOTE:** To authenticate in Kyma using GitHub, the user must be a member of a GitHub [organization](https://help.github.com/articles/creating-a-new-organization-from-scratch/) that has at least one [team](https://help.github.com/articles/creating-a-team/).


  </details>
  <details>
  <summary>
  XSUAA
  </summary>

  To add an XSUAA connector to Dex, register an OAuth2 client in SAP CP XSUAA. Set the authorization callback URL to `https://dex.{CLUSTER_DOMAIN}/callback`.

  </details>

</div>

## Configure Dex

Register the connector by creating a [Helm override](/docs/root/#configuration-helm-overrides-for-kyma-installation) for Dex. Create the override ConfigMap in the Kubernetes cluster before Dex is installed. If you want to register a connector at runtime, trigger the [update process](/docs/root/#installation-update-kyma-trigger-the-update-process) after creating the override.

>**TIP:** You can use Go Template expressions in the override value. These expressions are resolved by Helm using the same set of overrides as configured for the entire chart.

<div tabs>
  <details>
  <summary>
  GitHub
  </summary>

  ```bash
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
      "dex.useStaticConnector": "false"
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

  These are the placeholders used in the template:
  - GITHUB_CLIENT_ID - specifies the application's client ID.
  - GITHUB_CLIENT_SECRET - specifies the application's client Secret.
  - GITHUB_ORGANIZATION - specifies the name of the GitHub organization.


  </details>
  <details>
  <summary>
  XSUAA
  </summary>

  ```bash
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
      "dex.useStaticConnector": "false"
      connectors: |-
        - type: xsuaa
          id: xsuaa
          name: XSUAA
          config:
            issuer: {XSUAA_ISSUER}
            clientID: {XSUAA_OAUTH_CLIENT_ID}
            clientSecret: {XSUAA_OAUTH_CLIENT_SECRET}
            redirectURI: https://dex.{{ .Values.global.domainName }}/callback
            userNameKey: "{KEY_STRING}"
            appname: "{READABLE_APP_NAME}"
  EOF
  ```

  These are the placeholders used in the template:
  - XSUAA_OAUTH_CLIENT_ID - specifies the application's client ID.
  - XSUAA_ISSUER - specifies the XSUAA token issuer.
  - XSUAA_OAUTH_CLIENT_SECRET - specifies the application's client Secret.
  - KEY_STRING - specifies the string in the token that precedes the name of the user for which the token is issued.
  - READABLE_APP_NAME - specifies an additional, human-readable identifier for the OAuth2 client application.
  
  >**TIP:** The XSUAA connector supports Refresh Tokens. Include the `offline_access` scope in the authentication request to obtain an Access Token accompanied by a Refresh Token. Use the Refresh Token to renew expired Access Tokens. To revoke the Refresh Token, delete the corresponding instance of the `refreshtokens.dex.coreos.com` CR from the `kyma-system` Namespace. 

  </details>

</div>

>**TIP:** The **dex.useStaticConnector** parameter set to `false` overrides the default configuration and disables the Dex static user store. As a result, you can login to the cluster using only the registered connectors. If you want to keep the Dex static user store enabled, remove the **dex.useStaticConnector** parameter from the ConfigMap template.

## Configure authorization rules for the GitHub connector

To bind Github groups to the default Kyma roles, edit the **bindings** section in [this](https://github.com/kyma-project/kyma/blob/master/resources/core/charts/cluster-users/values.yaml) file. Follow this template:

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

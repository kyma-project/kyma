# Ory

[Ory](https://www.ory.sh/) Open Source OAuth 2.0 & OpenID Connect

## Introduction

This chart bootstraps [Hydra](https://www.ory.sh/docs/hydra/) and [Oathkeeper](https://www.ory.sh/docs/oathkeeper/) components on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Chart Details

This chart installs two Ory components as subcharts:
- hydra
- oathkeeper

To enable Hydra-Dex integration:

- Add new static client to Dex configuration in *resources/dex/templates/dex-config-map.yaml* file.
  Replace <domainName> with the proper domain name of your ingress gateway for cluster installations or *kyma.local* for local installations.
  Replace `<secretValue>` with a secure random string.

  ```
    - id: hydra-integration
      name: 'Hydra Integration'
      redirectURIs:
      - 'https://oauth2-login-consent.<domainName>/cb'
      secret: <secretValue>
  ```
  Alternatively, you can change this at runtime by modifying Dex configuration on a running installation and restarting Dex.

- Change hydra server configuration to point to a valid login-and-consent application.
  Modify *resources/ory/charts/hydra/values.yaml* file and set the `loginConsent.name` to the same value as defined in __this__ chart: `oauth2-login-consent`
  Alternatively, you can change this at runtime by changing the values of `OAUTH2_CONSENT_URL` and `OAUTH2_LOGIN_URL` environment variables of the Hydra server Deployment `ory-hydra-oauth2` in `kyma-system` namespace. Ensure hydra server Pod is redeployed with new values.

- Install Kyma with *ory* chart enabled.

- Use Helm to install hydra-dex integration chart.
  Replace <domainName> with the proper domain name of your ingress gateway for cluster installations or *kyma.local* for local installation.
  `helm install resources/hydra-dex -n hydra-dex --namespace kyma-system --set loginConsent.domainName=kyma.local --tls`

- Get Issuer URL from deployed application:
  `echo $(kubectl get deployments/ory-hydra-oauth2 -n kyma-system -o go-template='{{range (index .spec.template.spec.containers 0).env}}{{if eq .name "OAUTH2_ISSUER_URL"}}{{.value}}{{end}}{{end}}')
`
  Example output: `https://oauth2.kyma.local/`
- Create a lambda
  Expose the lambda as HTTPS.
  Set `Issuer` to Isuser URL from prevous step, Set `JWKS URI` to: `http://ory-hydra-oauth2.kyma-system.svc.cluster.local/.well-known/jwks.json`

- Add a client
  Replace <domainName> with the proper domain name of your ingress gateway for cluster installations or *kyma.local* for local installation.
  `curl -ik -X POST "https://oauth2-admin.<domainName>/clients" -d '{"grant_types":["implicit"], "response_types":["id_token"], "scope":"openid", "redirect_uris":["http://localhost:8080/callback"], "client_id":"implicit-client", "client_secret":"some-secret"}'`

- Fetch a JWT token

  Use your browser with the following URL:
  `"http://oauth2.<domainName>/oauth2/auth?client_id=implicit-client&response_type=id_token&scope=openid&state=8230b269ffa679e9c662cd10e1f1b145&redirect_uri=http://localhost:8080/callback&nonce=someNonce"`

- Call the lambda with the token

  `curl -ik -X GET "https://hydra-demo-production.<domainName>/" -H "Authorization: Bearer $JWT"`



# Eventing Chart

This Helm chart contains all components required for eventing in Kyma:

Components:
- event-publisher-proxy

## Event Publisher Proxy

This component receives Cloud Event publishing requests from the cluster workloads (microservice or Serverless functions) and redirects them to the Enterprise Messaging Service Cloud Event Gateway. Click [here](https://github.com/kyma-project/kyma/tree/master/components/event-publisher-proxy) for more details.

## Install

You can install this Helm chart using either Helm or Kyma CLI. In both cases, the secret details for BEB have to be configured using the `beb` prefixed variables:

### Using Helm 3:

```bash
$ cat << EOF > helm-values.yaml
event-publisher-proxy:
  upstreamAuthentication:
    oauthClientId: "$bebOauthClientId"
    oauthClientSecret: "$bebOauthClientSecret"
    oauthTokenEndpoint: "$bebOauthTokenEndpoint"
    publishUrl: "$bebPublishUrl"
EOF

$ helm install \
    -n kyma-system \
    -f helm-values.yaml \
     eventing . \
```

### Using Kyma CLI:

```bash
cat << EOF > installation-overrides-epp.yaml
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: eventing-epp-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: eventing
    kyma-project.io/installation: ""
stringData:
    event-publisher-proxy.upstreamAuthentication.oauthClientId: "$bebOauthClientId"
    event-publisher-proxy.upstreamAuthentication.oauthClientSecret: "$bebOauthClientSecret"
    event-publisher-proxy.upstreamAuthentication.oauthTokenEndpoint: "$bebOauthTokenEndpoint"
    event-publisher-proxy.upstreamAuthentication.publishUrl: "$bebPublishUrl"
EOF

$ kyma install -s <source-image> -o installation-overrides-epp.yaml
```

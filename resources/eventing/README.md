# Eventing Chart

This Helm chart contains all components required for eventing in Kyma:

Components:
- event-publisher-proxy
- controller

## Event Publisher Proxy

This component receives legacy and Cloud Event publishing requests from the cluster workloads (microservice or Serverless functions) and redirects them to the Enterprise Messaging Service Cloud Event Gateway. It also fetches a list of subscriptions for a connected application. Click [here](../../components/event-publisher-proxy) for more details.

## Controller

This component manages the internal infrastructure in order to receive an event for all subscriptions.

## Installation

You can install this Helm chart using either Helm or Kyma CLI. In both cases, the secret details for BEB have to be configured using the `beb` prefixed variables:

### Using Helm 3:


```bash
# Install subscriptions.eventing.kyma-project.io CRD
kubectl apply -f resources/cluster-essentials/files/subscriptions.eventing.kyma-project.io.crd.yaml

# Set values for chart
$ cat << EOF > helm-values.yaml
global:
  domainName: "$domainName"
authentication:
  oauthClientId: "$bebOauthClientId"
  oauthClientSecret: "$bebOauthClientSecret"
  oauthTokenEndpoint: "$bebOauthTokenEndpoint"
  publishUrl: "$bebPublishUrl"
  bebNamespace: "$bebNamespace"
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
  name: eventing-beb-auth-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: eventing
    kyma-project.io/installation: ""
stringData:
    authentication.oauthClientId: "$bebOauthClientId"
    authentication.oauthClientSecret: "$bebOauthClientSecret"
    authentication.oauthTokenEndpoint: "$bebOauthTokenEndpoint"
    authentication.publishUrl: "$bebPublishUrl"
    authentication.bebNamespace: "$bebNamespace"
EOF

$ kyma install -s <source-image> -o installation-overrides-epp.yaml
```

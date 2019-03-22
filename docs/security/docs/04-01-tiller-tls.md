---
title: TLS in Tiller
type: Installation
---

## Secured by default
Kyma is supplied with a custom installation of [Tiller](https://helm.sh/docs/glossary/#tiller), which secured all incoming connections with a TLS certificate. Because of that, all client connections (whether from inside or outside of the cluster) require a special pair of client certificates. 

## Retrieving 
In order to secure a local connection the certificates need to be downloaded from the cluster and stored in [`HELM_HOME`](https://helm.sh/docs/glossary/#helm-home-helm-home). This can be done by calling:

```
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.ca\.crt']}" | base64 -D > "$(helm home)/ca.pem"
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.crt']}" | base64 -D > "$(helm home)/cert.pem"
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.key']}" | base64 -D > "$(helm home)/key.pem"
```

Content of `HELM_HOME`
```
├── ca.pem              # <- Certificate Authority for Helm
├── cache
├── cert.pem            # <- Helm client certificate
├── key.pem             # <- Helm client key
├── plugins
├── repository
└── starters
```

> **NOTE:** By default those certificates are valid for 1 year. 

## Helm client usage
With the certs present in `HELM_HOME` a secured tls connection can be enabled by adding a `--tls` flag to a helm client call (for example `helm ls` becomes `helm ls --tls`).
If the flag is not added, or the certs are invalid, helm will return an error:
```
helm list 
Error: transport is closing
```

## Developer user guide
In order to connect to the Tiller server (for example using the [Helm GO library](https://godoc.org/k8s.io/helm/pkg/helm#pkg-index)) it is required to mount the helm client secrets into the application. For ease those certificates are available as a secret in kyma: 

```
kubectl get secret -n kyma-installer helm-secret                            
NAME          TYPE      DATA      AGE
helm-secret   Opaque    3         5m8s
```

Additionally, those secrets are also available as overrides during the kyma installation phase:

| Override | Description |
| :--- | --- | 
| global.helm.ca.crt | Certificate Authority for the Helm Client |
| global.helm.tls.crt | User certificate for the Helm client | 
| global.helm.tls.key | User key for the Helm Client | 


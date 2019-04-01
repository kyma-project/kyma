---
title: TLS in Tiller
type: Details
---

Kyma comes with a custom installation of [Tiller](https://helm.sh/docs/glossary/#tiller), which secures all incoming traffic with TLS certificate verification. To enable communication with Tiller, you must save the client certificate and key and the cluster Certificate Authority (CA) to [Helm Home](https://helm.sh/docs/glossary/#helm-home-helm-home). 

Saving the client certificate, key, and CA to [Helm Home](https://helm.sh/docs/glossary/#helm-home-helm-home) is manual on cluster deployments. When you install Kyma locally, this process is handled by the `run.sh` script. 

Additionally, you must add the `--tls` flag to every Helm command. 
If you don't save the required certificates in Helm Home, or you don't include the `--tls` flag when you run a Helm command, you get this error: 
```
Error: transport is closing
```

## Add certificates to Helm Home

To get the client certificate and key and the cluster CA and add them to [Helm Home](https://helm.sh/docs/glossary/#helm-home-helm-home), run these commands: 
  ```bash
  kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.ca\.crt']}" | base64 --decode > "$(helm home)/ca.pem"
  kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.crt']}" | base64 --decode > "$(helm home)/cert.pem"
  kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.key']}" | base64 --decode > "$(helm home)/key.pem"
  ```

> **NOTE:** By default, those certificates are valid for 1 year since Kyma is installed. 

## Development

To connect to the Tiller server, for example using the [Helm GO library](https://godoc.org/k8s.io/helm/pkg/helm#pkg-index), mount the Helm client certificates into the application you want to connect. These certificates are stored as a Kubernertes Secret. 

To get this Secret, run: 
  ```bash
  kubectl get secret -n kyma-installer helm-secret                            
  ```

Additionally, those secrets are also available as overrides during Kyma installation:

| Override | Description |
| --- | --- | 
| global.helm.ca.crt | Certificate Authority for the Helm client |
| global.helm.tls.crt | Client certificate for the Helm client | 
| global.helm.tls.key | Client certificate key for the Helm client | 


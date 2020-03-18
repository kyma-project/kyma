# API Server Proxy

## Overview

The Kyma API Server Proxy is a core component that uses JWT authentication to secure access to the Kubernetes API server. It is based on the [kube-rbac-proxy](https://github.com/brancz/kube-rbac-proxy) project. This [Helm chart](../../resources/apiserver-proxy/Chart.yaml) outlines the component's installation.

## Prerequisites

Use these tools to work with the API Server Proxy:

- [Go](https://golang.org)
- [Docker](https://www.docker.com/)

## Details

This section describes:

- How to run the controller locally
- How to build the Docker image for the production environment
- How to use the environment variables
- How to test the Kyma API Server Proxy

### Run the component locally

Run Minikube to use the API Server Proxy locally. Run this command to run the application without building the binary:

```bash
go run cmd/proxy/main.go
```

### API Server Proxy configuration

You can use command-line flags to configure the API Server Proxy. Use these flags to secure the Kubernetes API Server with JWT authentication:

```txt
	--upstream="https://kubernetes.default"		 	The upstream URL to proxy to once requests have successfully been authenticated and authorized.
	--oidc-issuer="https://dex.{{ DOMAIN }}"		The URL of the OpenID issuer, only HTTPS scheme will be accepted. If set, it will be used to verify the OIDC JSON Web Token (JWT).
	--oidc-clientID="								The client ID for the OpenID Connect client, must be set if oidc-issuer-url is set.
	--oidc-ca-file="path/to/cert/file"				If set, the OpenID server's certificate will be verified by one of the authorities in the oidc-ca-file, otherwise the host's root CA set will be used.
```

Find more details about available flags [here](https://github.com/brancz/kube-rbac-proxy/blob/master/README.md)

### Test

Run all tests:

```bash
go test -v ./...
```

# API Server Proxy

## Overview

The Kyma API Server Proxy is a core component that secure access to the kubernetes apiserver using JWT authentication. 
It is based on `kube-rbac-proxy`(https://github.com/brancz/kube-rbac-proxy) project.
This [Helm chart](/resources/core/charts/apiserver-proxy/Chart.yaml) defines the component's installation.

## Prerequisites

You need these tools to work with the API Server Proxy:

- [Go distribution](https://golang.org)
- [Docker](https://www.docker.com/)


## Details

This section describes how to run the controller locally, how to build the Docker image for the production environment, how to use the environment variables, and how to test the Kyma API Server Proxy.

### Run the component locally

Run Minikube to use the API Server Proxy locally. Run this command to run the application without building the binary:

```bash
$ go run cmd/proxy/main.go
```

### API Server Proxy configuration

Application can be configured using command line flags, in order to secure kubernetes apiserver with JWT authentication use these flags:

```txt
	--upstream="https://kubernetes.default"		 	The upstream URL to proxy to once requests have successfully been authenticated and authorized.
	--oidc-issuer="https://dex.{{ DOMAIN }}"		The URL of the OpenID issuer, only HTTPS scheme will be accepted. If set, it will be used to verify the OIDC JSON Web Token (JWT).
	--oidc-clientID="								The client ID for the OpenID Connect client, must be set if oidc-issuer-url is set.
	--oidc-ca-file="path/to/cert/file"				If set, the OpenID server's certificate will be verified by one of the authorities in the oidc-ca-file, otherwise the host's root CA set will be used.
```

More details about available flags can be found [here](https://github.com/brancz/kube-rbac-proxy/blob/master/README.md)

### Test

Run all tests:

```bash
$ go test -v ./...
```


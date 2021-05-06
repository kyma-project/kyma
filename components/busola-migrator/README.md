# Busola Migrator

## Overview

Busola Migrator is an HTTP server that serves a static webpage. It allows for migrating business user permissions from the SAP authentication and authorization service (XSUAA) to in-cluster roles. It also redirects the users that try to access the Console UI from the previous Kyma Console link to the new Busola URL.

## Prerequisites

To set up the project, download these tools:

* [Go](https://golang.org/dl/) v1.15 or higher
* [Docker](https://www.docker.com/) in the newest version

## Usage

To work correctly, the component requires a running Kubernetes cluster.
Before running the program, make sure your active Kubeconfig is pointing at the correct cluster.  

### Environment variables

| Variable | Description | Default value |
| --- | --- | --- |
| **APP_PORT** | Port on which the server is listening | `80` |
| **APP_DOMAIN** | Domain on which the server is running  | `localhost` |
| **APP_TIMEOUT_READ** | Maximum amount of time allowed for the client to read the entire request, including the body, before timing out | `30s` |
| **APP_TIMEOUT_WRITE** | Maximum amount of time allowed for the server to send a response before timing out | `30s` |
| **APP_TIMEOUT_IDLE** | Maximum amount of time allowed for the client to wait for the next request before timing out (with keep-alives enabled) | `120s` |
| **APP_BUSOLA_URL** | URL of the Busola cluster | `https://busola.main.hasselhoff.shoot.canary.k8s-hana.ondemand.com` |
| **APP_OIDC_ISSUER_URL** | OpenID Connect (OIDC) issuer URL | `https://kyma.accounts.ondemand.com` |
| **APP_OIDC_CLIENT_ID** | OIDC client ID | `6667a34d-2ea0-43fa-9b13-5ada316e5393` |
| **APP_OIDC_SCOPE** | OIDC scope | `openid` |
| **APP_OIDC_USE_PKCE** | Parameter specifying if OIDC should use Proof Key for Code Exchange (PKCE) | `false` |
| **APP_STATIC_FILES_DIR** | Directory to look for the static webpage to serve | `./static` |
| **OVERRIDE_BUSOLA_URL** | Optional override for the Busola cluster URL | None |
| **OVERRIDE_OIDC_ISSUER_URL** | Optional override for the OIDC issuer URL | None |
| **OVERRIDE_OIDC_CLIENT_ID** | Optional override for the OIDC client ID | None |
| **OVERRIDE_OIDC_SCOPE** | Optional override for the OIDC scope | None |
| **OVERRIDE_OIDC_USE_PKCE** | Optional override for the use of PKCE in OIDC  | None |
| **APP_UAA_URL** | User Account and Authentication (UAA) server URL | None |
| **APP_UAA_CLIENT_ID** | UAA client ID  | None |
| **APP_UAA_CLIENT_SECRET** | UAA client secret | None |

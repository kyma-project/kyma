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
|**APP_KUBECONFIG_ID**|Kubeconfig ID to be requested from Busola to access the cluster safely|None|
| **APP_BUSOLA_URL** | URL of the Busola cluster | `https://dashboard.dev.kyma.cloud.sap/` |
| **APP_STATIC_FILES_DIR** | Directory to look for the static webpage to serve | `./static` |
| **OVERRIDE_BUSOLA_URL** | Optional override for the Busola cluster URL | None |
| **APP_UAA_ENABLED** | Parameter specifying whether the User Account and Authentication (UAA) migration functionality is enabled | `true` |
| **APP_UAA_URL** | UAA server URL | None |
| **APP_UAA_CLIENT_ID** | UAA client ID  | None |
| **APP_UAA_CLIENT_SECRET** | UAA client secret | None |

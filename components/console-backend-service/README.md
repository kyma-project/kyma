# Console Backend Service

## Overview

This project includes a server that exposes the GraphQL API for all Kyma UIs. It consumes the Kubernetes API using the K8S Go client.
This document describes how to use the application and how to develop new features in this project.

> **NOTE:** The description of the application configuration, the project structure, the architecture, and other project-specific details are located in the [`docs`](./docs/README.md) directory.

See the [GraphQL schema definition](internal/gqlschema/schema.graphql) file for the list of supported queries and mutations.

## Prerequisites

Use the following tools to set up the project:

* [Go](https://golang.org)
* [Docker](https://www.docker.com/)

## Usage

### Run a local version

To run the application without building the binary, run this command:

```bash
APP_KUBECONFIG_PATH=/Users/$USER/.kube/config APP_VERBOSE=true APP_RAFTER_ADDRESS="rafter-minio.kyma-system.svc.cluster.local:9000" APP_RAFTER_VERIFY_SSL=false APP_APPLICATION_GATEWAY_INTEGRATION_NAMESPACE=kyma-integration APP_APPLICATION_CONNECTOR_URL=http://dummy.url APP_OIDC_ISSUER_URL=https://dex.{kymaDomain} APP_OIDC_CLIENT_ID=kyma-client go run main.go
```

For the descriptions of the available environment variables, see the [Configuration](./docs/configuration.md) document.

The service listens on port `3000`. Open `http://localhost:3000` to see the GraphQL console in your browser.

### Use GraphQL console on cluster

Before using the console on a cluster, set a valid token for all requests. Click the **Header** option at the bottom of the GraphQL console and paste this snippet:

```json
{
    "Authorization": "Bearer {YOUR_BEARER_TOKEN}"
}
```

After you paste the custom HTTP header, reload the page. GraphQL console allows you to access the schema documentation and test queries, mutations, and subscriptions.

### Build a production version

To build the production Docker image, run this command:

```bash
docker build {image_name}:{image_tag}
```

The variables are:

* `{image_name}` - name of the output image (default: `console-backend-service`)
* `{image_tag}` - tag of the output image (default: `latest`)

### Certificate error

When you run the UI API Layer project, you can get the following error:

```bash
oidc.go:222] oidc authenticator: initializing plugin: Get https://dex.kyma.local/.well-known/openid-configuration: x509: certificate signed by unknown authority
```

This error can occur if you use Go version 1.11.5 or lower on macOS. Try upgrading to version 1.11.6 or higher. For details, see [this](https://github.com/golang/go/issues/24652) issue.

## Development

### Install dependencies

This project uses `dep` as a dependency manager. To install all required dependencies, use the following command:

```bash
dep ensure -vendor-only
```

#### Generate code from GraphQL schema

This project uses the [GQLGen](https://github.com/99designs/gqlgen) library, which improves development by generating code from the [GraphQL schema definition](internal/gqlschema/schema.graphql).

1. Define types and their fields in `/internal/gqlschema/schema.graphql` using the [Schema Definition Language](http://graphql.org/learn/schema/).
1. Execute the `./gqlgen.sh` script to run the code generator.
1. Navigate to the `/internal/gqlschema/` directory.
1. Find newly generated methods in the `ResolverRoot` interface located in `./schema_gen.go`.
1. Implement resolvers in specific domains according to the project structure and rules in this guide. Use generated models from `./models_gen.go` in your business logic. If you want to customize them, move them to a new file in the `gqlschema` package and include in the `./config.yml` file.

To use advanced features, such as custom scalars, read the [documentation](https://gqlgen.com/) of the used library.

### Run tests

To run all unit tests, execute the following command:

```bash
go test ./...
```

### Verify the code

To check if the code is correct and you can push it, use the `make` command. It builds the application, runs tests, checks the status of the vendored libraries, runs the static code analysis, and checks if the formatting of the code is correct. 

To automatically format the incorrect code, use the `make format` command.

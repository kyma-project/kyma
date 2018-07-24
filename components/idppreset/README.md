# IDP Preset

This repo contains generated client with `clientset`, `informers` and `listers`, thanks to which we are able to manage (create/delete/get) IDP Preset resources from k8s.

## Overview

An identity provider is a trusted provider that lets us use a single sign-on (SSO).
When exposing an API in the [Console](https://github.com/kyma-project/console-new/core), a user is able to secure it by loading a predefined IDP preset and use the issuer, jwksUri values from the selected preset.

## Prerequisites

You need the following tools to set up the project:
* The 1.9 or higher version of [Go](https://golang.org/dl/)
* The latest version of [Docker](https://www.docker.com/)
* The latest version of [Dep](https://github.com/golang/dep)

## Development

Install all dependencies:
```bash
dep ensure -vendor-only
```

Before each commit, use the `before-commit.sh` script, which tests your changes.

## Code generation

Structs related to Custom Resource Definitions are defined in `pkg/apis/ui/v1alpha1/types.go` and registered in `pkg/apis/ui/v1alpha1/`. After making any changes there, please run:
```bash
./hack/update-codegen.sh
```

## Release

Please, make sure that your _master_ branch is up-to-date with the changes that you want to release.

To release a new version, run on _master_ branch:
```bash
git tag YOUR_NEW_VERSION
git push --tags
```

##### IDPPreset as a ui-api-layer dependency

After releasing a new version in **idppreset** repository, change the version in Gopkg.toml in [ui-api-layer](https://github.com/kyma-project/kyma/components/ui-api-layer) project:
```
# Gopkg.toml file in ui-api-layer project
...
[[constraint]]
  name = "github.com/kyma-project/idppreset"
  version = "YOUR_NEW_VERSION"
...
```

# Microfrontend

This repo contains generated client with `clientset`, `informers` and `listers`, thanks to which we are able to manage (create/delete/get) Microfrontend resources from k8s.

## Overview

A micro frontend is a standalone web application which extends the default functionalities of the Kyma Console UI, but it is developed, tested and deployed independently from it.

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

Structs related to CustomResourceDefinitions are defined in `pkg/apis/ui/v1alpha1/types.go` and registered in `pkg/apis/ui/v1alpha1/`. After making any changes there, please run:
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


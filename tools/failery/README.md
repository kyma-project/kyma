# Failery

## Overview

This project allows you to generate failing mock types for given interfaces. It is based on the [mockery](https://github.com/vektra/mockery) project.

A failing mock type is a type that implements an interface. In case any method returns an error or error pointer, a failing mock type also returns an error.

## Prerequisites

Use the following tools to set up the project:

- [Go](https://golang.org)
- [dep package manager](https://github.com/golang/dep)

## Installation

To install this tool, run the following command:

```bash
dep ensure -add 'github.com/kyma-project/kyma/tools/failery'
```

## Usage

1. Create an interface for which you want to generate a mock implementation type. For example:

    ```go
    type Requester interface {
        Get(path string) (string, error)
    }
    ```

1. Add the following line above your interface:

    ```go
    //go:generate go run {relativeVendorPath}/cmd/failery/failery.go -name=Requester -case=underscore {generationTypeParams}
    type Requester interface {
        Get(path string) (string, error)
    }
    ```

    Make sure that the `-name` parameter value is equal to the name of the interface for which you want to generate the mock implementation.

    Replace values in curly braces with the proper details, where:

    - `{relativeVendorPath}` is the relative path to the `vendor` directory of the project.
    - `{generationTypeParams}` are additional parameters that specify a mock generation type. The **-inpkg** parameter creates a file in the same package, and the **-output {packageName}** parameter generates a mock in the `{packageName}` package.

    For more configuration options, see the [mockery](https://github.com/vektra/mockery) documentation.

1. Run the following command in your project root:

    ```bash
    go generate ./...
    ```

1. See the generated type:

    ```go
    type Requester struct {
        err error
    }

    // NewRequester creates a new Requester type instance
    func NewRequester(err error) *Requester {
        return &Requester{err: err}
    }

    // Get provides a failing mock Function with given fields: path
    func (_m *Requester) Get(path string) (string, error) {
        var r0 string
        var r1 error
        r1 = _m.err

        return r0, r1
    }
    ```

## Development

### Install dependencies

This project uses `dep` as a dependency manager. To install all required dependencies, use the following command:

```bash
dep ensure -vendor-only
```

### Verify the code

To check if the code is correct and you can push it, run the `before-commit.sh` script. It builds the application, runs tests, checks the status of the vendored libraries, runs the static code analysis, and ensures that the formatting of the code is correct.

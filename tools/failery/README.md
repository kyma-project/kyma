# Failery

## Overview
This project allows to generate failing mock types for given interfaces.

What is a failing mock type? It is a type that implements an interface, but when any of method returns an error or pointer to it, it will always return an error.

Failery is based on [failery](https://github.com/vektra/failery).

## Prerequisites

Use the following tools to set up the project:

* [Go distribution](https://golang.org)
* [dep package manager](https://github.com/golang/dep)

## Installation

To install this tool, run the following command:

```bash
dep ensure -add 'github.com/kyma-project/kyma/tools/failery'
```

## Usage

1. Create an interface, for which you want to generate mock implementation type. For example:

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
    
    Make sure that the `-name` parameter value is equal to the name of the interface, for which you want to generate the mock implementation.
    
    Replace values in curly braces with proper details, where:
    - `{relativeVendorPath}` is the relative path to the `vendor` directory of the project.
    - `{generationTypeParams}` are additional parameters that specify mock generation type. The `-inpkg` parameter creates file in the same package, and `-output {packageName}` parameter generates the mock in `{packageName}` package.
    
    For more configuration options, see the Readme of the original project, [failery](https://github.com/vektra/failery).
  
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
    
    // Get provides a failing mock function with given fields: path
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

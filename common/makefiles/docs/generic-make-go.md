# Generic Makefile
## Overview

`generic_make_go.mk` is a generic Makefile used to build, format, and test Golang components.

## Prerequisites 
* [Docker](https://www.docker.com/get-started)
* [GNU Makefile](https://www.gnu.org/software/make/manual/make.html) v3.80 or higher

For local usage, you need:
* [Go](https://golang.org/) v1.11 or higher
* [errcheck](https://github.com/kisielk/errcheck)
* [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports)
* [dep](https://github.com/golang/dep) v0.5.1 or higher
* [Mockery](https://github.com/vektra/mockery)

## Add the generic Makefile
To add the generic Makefile to your Makefile, add these variables and the statement:
```makefile
APP_NAME = {APPLICATION NAME}
APP_PATH = {APPLICATION PATH IN REPOSITORY}
BUILDPACK = {BUILDPACK IMAGE}
SCRIPTS_DIR = {GENERIC MAKEFILE PATH} 
include $(SCRIPTS_DIR)/generic-make-go.mk
```
Find the list of available buildpack images [here](https://github.com/kyma-project/test-infra/blob/master/templates/config.yaml).

## Usage
This is the basic syntax used in Makefiles:
```bash
make {RULE}
```

The default rule in the generic Makefile is `verify`, which means that when you run `make` without specifying any rule, the `verify` rule is executed.
It runs tests, checks formatting and imports, and runs error checks.

These are the possible rules that you can use:
>**NOTE:** Rules without the `-local` suffix are used inside the Docker container.
The Docker container is used in the CI environment.

| Rule                              | Description                                                    |
|-----------------------------------|----------------------------------------------------------------|
| check-fmt, check-fmt-local        | Check the formatting of Go files.                              |
| fmt, fmt-local                    | Format Go files.                                               |
| check-imports, check-imports-local| Check Go files imports.                                        |
| imports, imports-local            | Format Go imports.                                             |
| gqlgen, gqlgen-local              | Generate GraphQL schema.                                       |
| check-gqlgen, check-gqlgen-local  | Check if GraphQL schema is up-to-date. Use it after the `gqlgen` rule. |
| errcheck, errcheck-local          | Run the [errcheck](https://github.com/kisielk/errcheck) program.        |
| test, test-local                  | Run all unit tests.                                            |
| verify                            | Run `test`,`check-fmt`, `check-imports`, and `errcheck`.       |
| resolve, resolve-local            | Run the `dep resolve --vendor-only -v` command which installs dependencies defined in the `Gopkg.lock` file without updating the file itself.                         |
| ensure, ensure-local              | Run the `dep ensure -v` command which downloads dependencies defined in the `Gopkg.lock` file and updates the file itself.                               |
| dep-status, dep-status-local      | Run the `dep status -v` command which prints the status of project dependencies.                                         |
| build, build-local                | Build Go binary.                                               |
| vet, vet-local                    | Run the `go vet` command which examines Go source code.                                      |
| build-image                       | Builds a Docker image used in the CI environment.                 |
| push-image                        | Pushes the image to the image registry used in the CI environment.           |

### Example workflow 
The generic Makefile is used by the CI.
When a job is triggered, the CI runs the `make release` command.
The `release` rule depends on the `resolve`, `dep-status`, `verify`, `build-image`, and `push-image` rules.
Although these rules do not appear in the Makefile, they are generated. 
 
The generic Makefile contains such a line of code:
```makefile
MOUNT_TARGETS = build resolve ensure dep-status check-imports imports check-fmt fmt errcheck vet generate pull-licenses gqlgen
$(foreach t,$(MOUNT_TARGETS),$(eval $(call buildpack-mount,$(t))))
```
For all the rules defined in **MOUNT_TARGETS**, the `buildpack-mount` function is called. It dynamically defines a new rule:
```makefile
resolve:
    @echo make resolve
    @docker run $(DOCKER_INTERACTIVE) \
        -v $(COMPONENT_DIR):$(WORKSPACE_COMPONENT_DIR):delegated \
        $(DOCKER_CREATE_OPTS) make resolve-local
```
As you can see, the `resolve-local` rule is executed inside the container. 
If the `resolve` rule passes, the next rules are executed.
If any rule fails, the Makefile aborts the execution and fails as well.

There are two available **BUILDPACK_FUNCTIONS** which generate the rule dynamically:
- `buildpack-mount` - creates the rule and mounts component's directory as volume
- `buildpack-cp-ro` - creates the rule and copies component's files to a Docker container

These are two variables which contain rules names that are created dynamically:
- **MOUNT_TARGET** - contains rules which are dynamically created by the `buildpack-mount` function
- **COPY_TARGET** - contains rules which are dynamically created by the `buildpack-cp-ro` function

## Adjust the generic Makefile

### Disable the current rule in the local Makefile
To disable a rule in the new Makefile, follow it with the semicolon `;`.
For example, write: `{RULE}: ;`.
The warning will appear in the console but the rule will be disabled.

### Add a new local rule that doesn't need a buildpack
To add a new local rule that doesn't need a buildpack, define a rule in the local Makefile and add this rule to one of the global rule:
```makefile
verify:: {YOUR_RULE}
```

### Add a new local rule that needs a buildpack
To add a new local rule that needs a buildpack, define a rule in the local Makefile and call the function that creates the rule:
```makefile
{RULE}-local: {COMMANDS}

# Function which creates the new rule dynamically
$(eval $(call {BUILDPACK_FUNCTION},{RULE})) 
```

### Add a new rule in the generic Makefile
To add a new rule in the generic Makefile, define a new local rule in the `generic_make_go.mk` file:
```makefile
{YOUR_RULE}-local:  {COMMANDS}
```
Then, add {YOUR_RULE} to the **MOUNT_TARGETS** or **COPY_TARGETS** variables.

### Change artifacts directory

By default test coverage report is saved to `/tmp/artifacts` directory.
To change this location set the `ARTIFACTS` environment variable.
It is automatically set when running in [Prow environment](https://github.com/kubernetes/test-infra/blob/master/prow/pod-utilities.md)

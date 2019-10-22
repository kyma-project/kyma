# Generic Makefile
## Overview

`generic_make_go.mk` is a generic Makefile used to build, format and test Golang components.

## Prerequisites 
* [Docker](https://www.docker.com/get-started)
* [GNU Makefile](https://www.gnu.org/software/make/manual/make.html) v3.80 or higher

Local usage:
* [go](https://golang.org/) v1.11 or higher
* [errcheck](https://github.com/kisielk/errcheck)
* [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports)
* [dep](https://github.com/golang/dep) v0.5.1 or higher
* [Mockery](github.com/vektra/mockery)
## Usage
This is the basic syntax used in the Generic Makefile:
```bash
make {RULE}
```

The default rule is `verify`, which means that when you run `make`  without specifying any rule, the `verify` rule is executed.
It runs tests, checks formatting and imports, and runs error checks.`verify` rule will run tests, check formatting, check imports and run errcheck.

These are the possible rules that you can use:
>**NOTE:** Rules without the `-local` suffix are run inside a Docker container and are used in the CI environment.

| Rule                              | Description                                                    |
|-----------------------------------|----------------------------------------------------------------|
| check-fmt, check-fmt-local        | Check the formatting of Go files.                              |
| fmt, fmt-local                    | Format Go files.                                               |
| check-imports, check-imports-local| Check Go files imports.                                        |
| imports, imports-local            | Format Go imports.                                             |
| gqlgen, gqlgen-local              | Generate GraphQL schema.                                       |
| check-gqlgen, check-gqlgen-local  | Check if GraphQL schema is up-to-date. Use it after the `gqlgen` rule. |
| errcheck, errcheck-local          | Run the [errcheck](https://github.com/kisielk/errcheck) program        |
| test, test-local                  | Run all unit tests.                                            |
| verify                            | Run `test`,`check-fmt`, `check-imports`, and `errcheck`.       |
| resolve, resolve-local            | Run the `dep resolve --vendor-only -v` command.                         |
| ensure, ensure-local              | Run the `	dep ensure -v` command which download dependcies .                                           |
| dep-status, dep-status-local      | Run the `dep status -v` command which prints the status of project dependencies.                                         |
| build, build-local                | Build Go binary.                                               |
| vet, vet-local                    | Run the `go vet` command.                                      |
| build-image                       | Build a Docker image used in the CI environment.    ??                 |
| push-image                        | Pushes the image to image registry, used in CI environment.           |

### Add Generic Makefile
To add generic Makefile, add to your Makefile following things and fill the variables.
```makefile
APP_NAME = {APPLICATION NAME}
APP_PATH = {APPLICATION PATH IN REPOSITORY}
BUILDPACK = {BUILDPACK IMAGE}
SCRIPTS_DIR = {GENERIC MAKEFILE PATH} 
include $(SCRIPTS_DIR)/generic-make-go.mk
```
Find the list of available images [here](https://github.com/kyma-project/test-infra/blob/master/templates/config.yaml).

### Detailed workflow 
By example:
When CI run`make release` the following steps are executed:
- rule `release` depends on rules `resolve dep-status verify build-image push-image`
- rule`resolve` does not appear in the Makefile, but it's generated. 
Notice this line of code:
```makefile
MOUNT_TARGETS = build resolve ensure dep-status check-imports imports check-fmt fmt errcheck vet generate pull-licenses gqlgen
$(foreach t,$(MOUNT_TARGETS),$(eval $(call buildpack-mount,$(t))))
```
for resolve, and others targets defined in `MOUNT_TARGETS`, function `buildpack-mount` is callled, which dynamically defines new rule:
```makefile
resolve:
    @echo make resolve
    @docker run $(DOCKER_INTERACTIVE) \
        -v $(COMPONENT_DIR):$(WORKSPACE_COMPONENT_DIR):delegated \
        $(DOCKER_CREATE_OPTS) make resolve-local
```
  as you can see, rule `resolve-local` is executed inside container. 
- if `resolve` pass, then the next rule will be executed.
- if exit code of any rules is different than 0, makefile will abort execution and fail.

List of available BUILDPACK_FUNCTIONS, which generates target dynamically:
- buildpack-mount - create rule and mount component directory as volume
- buildpack-cp-ro - create rule and copy components files to docker container.

Target types:
- `MOUNT_TARGET` - Contains rules which will be dynamically created by `buildpack-mount` function.
- `COPY_TARGET` - Contains rules which will be dynamically created by `buildpack-cp-ro` function.

## How to adjust makefile
### Disable the current rule in the local Makefile
To disable a rule in the new Makefile, follow it with the semicolon `;`.
For example, write: `{RULE}: ;`.
This results in the rule being disabled and warnings printed on the console.
### How to add new local rule, which doesn't need `BUILDPACK`:
Define rule in local makefile.
Add this rule to one of the  global rule:
```makefile
verify:: own-rule
```

### How to add new rule in local makefile, which needs buildpack:
Define rule in local makefile and call function which will create the rule:
```makefile
{my-rule}-local: 
    do sth

$(eval $(call {BUILDPACK_FUNCTION},my-rule)) # function which will create the new rule
```

### Add a new rule in the Generic Makefile
Definie new local rule in `generic_make_go.mk` file:
```makefile
your-rule-local:
    @echo do sth
```

Append rule name to the `MOUNT_TARGETS` or `COPY_TARGETS` variables.

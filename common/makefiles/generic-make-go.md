# Generic Makefile
## Overview

`generic_make_go.mk` is a generic make file used to build, format and test Golang components.

## Prerequisites 
* [Docker](https://www.docker.com/get-started)
* [GNU Makefile](https://www.gnu.org/software/make/manual/make.html) v3.80 or higher

## Usage
Syntax:
```bash
make 'rule'
```

Default makefile rule is `verify`, so when `make` is launched, the `verify` rule will be executed.
`verify` rule will run tests, check formatting, check imports and run errcheck.
It can be used instead of old `/before-commit.sh`.

On CI all rules are run inside docker, because in CI environment we don't have `go` tools.

## Rules
Rules without ending `-local` are run inside docker container.

| RULE                              | BEHAVIOUR                                                      |
|-----------------------------------|----------------------------------------------------------------|
| check-fmt, check-fmt-local        | check `Go` files formatting                                    |
| fmt, fmt-local                    | format `Go` files                                              |
| check-imports, check-imports-local| check `Go` files imports                                       |
| imports, imports-local            | format `Go` imports                                            |
| gqlgen, gqlgen-local              | generate GraphQL schema                                        |
| check-gqlgen, check-gqlgen-local  | check if graphql schema is up to date, run after `gqlgen` rule |
| errcheck, errcheck-local          | run [errcheck](https://github.com/kisielk/errcheck)            |
| test, test-local                  | run all unit tests                                             |
| verify                            |  run `test`,`check-fmt`, `check-imports` and `errcheck`        |
| resolve, resolve-local            | run `dep resolve --vendor-only -v                              |
| ensure, ensure-local              | run `	dep ensure -v`                                           |
| dep-status, dep-status-local      | run 	`dep status -v`                                          |
| build, build-local                | build Go binary                                                |
| vet, vet-local                    | run `go vet`                                                   |
| build-image                       | build Docker image, used in CI environment                     |
| push-image                        | push image to Image registry, used in CI environment           |

## How to use `generic_make_go.mk` in your application makefile
Makefile must contains following vars and statement:
```makefile
APP_NAME = App name
APP_PATH = App path in repository
BUILDPACK = BUILDPACK_IMAGE 
SCRIPTS_DIR = Path to generic makefile #e.g. $(realpath $(shell pwd)/../..)/common/makefiles
include $(SCRIPTS_DIR)/generic-make-go.mk
```
available images are listed in [config.yaml](https://github.com/kyma-project/test-infra/blob/master/templates/config.yaml).

## How it works
By example:
When CI run`make release` the following steps are executed:
- rule `release` depends on rules `resolve dep-status verify build-image push-image`
- rule`resolve` does not appear in the Makefile. but it's generated. 
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

Target types:
- `MOUNT_TARGET` - mount component directory as volume to docker container
- `COPY_TAGRT` - copy components files to docker container.

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
my-rule-local: 
    do sth

$(eval $(call BUILDPACK_FUNCTION,my-rule)) # function which will create the new rule
```

Available BUILDPACK_FUNCTIONS:
- buildpack-mount
- buildpkac-cp-ro

### Add a new rule in the Generic Makefile
Definie new local rule in `generic_make_go.mk` file:
```makefile
your-rule-local:
    @echo do sth
```

Append rule name to the `MOUNT_TARGETS` or `COPY_TARGETS` variables.

The `kubeless-test-client` contains the following:
* A `Dockerfile` for the image used in Kyma kubeless tests
* A go file named test-kubeless.go, which executes the tests for the kubeless chart

Use the following command to build image:

```bash
$ make build-release-image RELEASE_VERSION=1.6.0
```

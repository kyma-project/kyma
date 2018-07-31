The `kubeless-test-client` contains the following:
* A [Dockerfile](Dockerfile) for the image used in Kyma Kubeless tests
* A [Go program](test-kubeless.go), which executes the tests for the Kubeless chart
* The [ns.yaml](ns.yaml) file, which specifies the `kubeless-test` namespace
* The [k8s.yaml](k8s.yaml) file, which contains necessary resources for `test-event` and `test-hello`
* The JavaScript files of `test-event` and `test-hello`
* The [svcbind.yaml](svcbind.yaml) file, which specifies `test-svcbind` and its supporting Kubernetes resources, including a Redis service which the functions binds to.

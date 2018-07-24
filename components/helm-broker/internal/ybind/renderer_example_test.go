//+build integration

package ybind_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	yhelm "github.com/kyma-project/kyma/components/helm-broker/internal/helm"
	"github.com/kyma-project/kyma/components/helm-broker/internal/platform/logger/spy"
	"github.com/kyma-project/kyma/components/helm-broker/internal/ybind"
	"k8s.io/helm/pkg/chartutil"
)

// To run it you need expose tiller pod port by:
// kubectl port-forward <tiller_pod> 44134:44134 -n kube-system
func ExampleNewRenderer() {
	const releaseName = "example-renderer-test"
	bindTmplRenderer := ybind.NewRenderer()

	// loadChart
	ch, err := chartutil.Load("testdata/repository/redis-0.0.3/chart/redis")
	fatalOnErr(err)

	// load bind template for above chart
	b, err := ioutil.ReadFile(filepath.Join("testdata/repository", "redis-0.0.3/plans/micro/bind.yaml"))
	fatalOnErr(err)

	hClient := yhelm.NewClient("localhost:44134", spy.NewLogDummy())

	// install chart in same way as we are doing in our business logic
	resp, err := hClient.Install(ch, internal.ChartValues{}, releaseName, "ns-name")

	// clean-up, even if install error occurred
	defer hClient.Delete(releaseName)
	fatalOnErr(err)

	rendered, err := bindTmplRenderer.Render(internal.BundlePlanBindTemplate(b), resp)
	fatalOnErr(err)

	fmt.Println(string(rendered))

	// Output:
	// credential:
	//   - name: HOST
	//     value: example-renderer-test-redis.ns-name.svc.cluster.local
	//   - name: PORT
	//     valueFrom:
	//       serviceRef:
	//         name: example-renderer-test-redis
	//         jsonpath: '{ .spec.ports[?(@.name=="redis")].port }'
	//   - name: REDIS_PASSWORD
	//     valueFrom:
	//       secretKeyRef:
	//         name: example-renderer-test-redis
	//         key: redis-password
}

func fatalOnErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

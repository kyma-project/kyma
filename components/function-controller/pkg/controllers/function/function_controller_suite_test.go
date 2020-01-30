/*
Copyright 2019 The Kyma Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package function

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	gm "github.com/onsi/gomega"
	tektonv1alpha2 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha2"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"

	"github.com/kyma-project/kyma/components/function-controller/pkg/apis"
	"github.com/kyma-project/kyma/components/function-controller/test"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var cfg *rest.Config

func TestMain(m *testing.M) {
	crdBases := filepath.Join("..", "..", "..", "config", "crd", "bases")
	crdThirdParty := filepath.Join("..", "..", "..", "config", "crd", "crds-thirdparty")
	test.FileExists(crdBases)
	test.FileExists(crdThirdParty)

	t := &envtest.Environment{
		Config: cfg,
		CRDDirectoryPaths: []string{
			crdBases, crdThirdParty,
		},
	}

	logf.SetLogger(logf.ZapLogger(false))

	if err := apis.AddToScheme(scheme.Scheme); err != nil {
		log.Error(err, "unable add serverless APIs to scheme")
		os.Exit(1)
	}

	if err := servingv1alpha1.SchemeBuilder.AddToScheme(scheme.Scheme); err != nil {
		log.Error(err, "unable add Knative Serving APIs to scheme")
		os.Exit(1)
	}

	if err := tektonv1alpha2.AddToScheme(scheme.Scheme); err != nil {
		log.Error(err, "unable add Tekton APIs to scheme")
		os.Exit(1)
	}

	var err error
	if cfg, err = t.Start(); err != nil {
		log.Error(err, "failed to start Kubernetes test environment")
		os.Exit(1)
	}

	code := m.Run()
	t.Stop()
	os.Exit(code)
}

// SetupTestReconcile returns a reconcile.Reconcile implementation that delegates to inner and
// writes the request to requests after Reconcile is finished. If the reconcile function encounters any error, it is written to the errors channel
func SetupTestReconcile(inner reconcile.Reconciler) (reconcile.Reconciler, chan reconcile.Request, chan error) {
	requests := make(chan reconcile.Request)
	errors := make(chan error)

	fn := reconcile.Func(func(req reconcile.Request) (reconcile.Result, error) {
		result, err := inner.Reconcile(req)
		if err != nil {
			fmt.Printf("Reconciler encountered error: %v", err)
			errors <- err
		}
		requests <- req
		return result, err
	})
	return fn, requests, errors
}

// StartTestManager adds recFn
func StartTestManager(mgr manager.Manager, g *gm.GomegaWithT) (chan struct{}, *sync.WaitGroup) {
	stop := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		g.Expect(mgr.Start(stop)).NotTo(gm.HaveOccurred())
	}()
	return stop, wg
}

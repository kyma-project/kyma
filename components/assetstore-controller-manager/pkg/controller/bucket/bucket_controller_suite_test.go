package bucket

import (
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/buckethandler"
	"github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	stdlog "log"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sync"
	"testing"
	"time"
)

var cfg *rest.Config

func TestMain(m *testing.M) {
	t := &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "..", "config", "crds")},
	}
	apis.AddToScheme(scheme.Scheme)

	var err error
	if cfg, err = t.Start(); err != nil {
		stdlog.Fatal(err)
	}

	code := m.Run()
	t.Stop()
	os.Exit(code)
}

// SetupTestReconcile returns a reconcile.Reconcile implementation that delegates to inner and
// writes the request to requests after Reconcile is finished.
func SetupTestReconcile(inner reconcile.Reconciler) (reconcile.Reconciler, chan reconcile.Request) {
	requests := make(chan reconcile.Request)
	fn := reconcile.Func(func(req reconcile.Request) (reconcile.Result, error) {
		result, err := inner.Reconcile(req)
		requests <- req
		return result, err
	})
	return fn, requests
}

// StartTestManager adds recFn
func StartTestManager(mgr manager.Manager, g *gomega.GomegaWithT) (chan struct{}, *sync.WaitGroup) {
	stop := make(chan struct{})
	wg := &sync.WaitGroup{}
	go func() {
		wg.Add(1)
		g.Expect(mgr.Start(stop)).NotTo(gomega.HaveOccurred())
		wg.Done()
	}()
	return stop, wg
}

type testSuite struct {
	g          *gomega.GomegaWithT
	c          client.Client
	mgr        manager.Manager
	stopMgr    chan struct{}
	mgrStopped *sync.WaitGroup
	requests   chan reconcile.Request

	finishTest func()
}

func prepareReconcilerTest(t *testing.T, handler buckethandler.BucketHandler) *testSuite {
	g := gomega.NewGomegaWithT(t)
	mgr, err := manager.New(cfg, manager.Options{})
	c := mgr.GetClient()

	g.Expect(err).NotTo(gomega.HaveOccurred())

	reconciler, err := newReconciler(mgr, handler, 60*time.Hour)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	recFn, requests := SetupTestReconcile(reconciler)
	g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())

	stopMgr, mgrStopped := StartTestManager(mgr, g)

	return &testSuite{
		g:        g,
		c:        c,
		requests: requests,
		finishTest: func() {
			close(stopMgr)
			mgrStopped.Wait()
		},
	}
}

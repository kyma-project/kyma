package knative

import (
	"context"
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/vrischmann/envconfig"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var config ServiceConfig
var k8sClient client.Client
var testEnv *envtest.Environment
var cfg *rest.Config

func TestAPIs(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)

	ginkgo.RunSpecsWithDefaultAndCustomReporters(t,
		"Serverless Suite - Knative Service Controller",
		[]ginkgo.Reporter{printer.NewlineReporter{}})
}

var _ = ginkgo.BeforeSuite(func(done ginkgo.Done) {
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(ginkgo.GinkgoWriter)))
	ginkgo.By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "..", "config", "crd", "bases"),
			filepath.Join("..", "..", "..", "config", "crd", "crds-thirdparty"),
		},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	cfg, err = testEnv.Start()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(cfg).ToNot(gomega.BeNil())

	err = scheme.AddToScheme(scheme.Scheme)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	err = servingv1.AddToScheme(scheme.Scheme)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(k8sClient).ToNot(gomega.BeNil())

	err = envconfig.InitWithPrefix(&config, "TEST")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	close(done)
}, 60)

var _ = ginkgo.AfterSuite(func() {
	ginkgo.By("tearing down the test environment")
	err := testEnv.Stop()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
})

// SetupTest will set up a testing environment.
// This includes:
// * creating a Namespace to be used during the test
// * starting reconciler
// * stopping the reconciler after the test ends
// Call this function at the start of each of your tests.
func SetupTest(ctx context.Context) *corev1.Namespace {
	var stopCh chan struct{}
	ns := &corev1.Namespace{}

	ginkgo.BeforeEach(func() {
		stopCh = make(chan struct{})
		*ns = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: "testns-" + randStringRunes(5)},
		}

		err := k8sClient.Create(ctx, ns)
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "failed to create test namespace")

		mgr, err := ctrl.NewManager(cfg, ctrl.Options{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "failed to create manager")

		reconcilerCfg := ServiceConfig{}
		err = envconfig.InitWithPrefix(&reconcilerCfg, "test")
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "failed to create controller config")

		controller := NewServiceReconciler(mgr.GetClient(), logf.Log, reconcilerCfg, scheme.Scheme, record.NewFakeRecorder(100))

		err = controller.SetupWithManager(mgr)
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "failed to setup controller")

		go func() {
			err := mgr.Start(stopCh)
			gomega.Expect(err).NotTo(gomega.HaveOccurred(), "failed to start manager")
		}()
	})

	ginkgo.AfterEach(func() {
		close(stopCh)

		err := k8sClient.Delete(ctx, ns)
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "failed to delete test namespace")
	})

	return ns
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

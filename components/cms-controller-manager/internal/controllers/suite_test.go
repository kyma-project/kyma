package controllers

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/cms-controller-manager/internal/webhookconfig"
	cmsv1alpha1 "github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var webhookConfigSvc webhookconfig.AssetWebhookConfigService
var fixParameters = runtime.RawExtension{Raw: []byte(`{"json":"true","complex":{"data":"true"}}`)}

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "config", "crd", "bases"),
			filepath.Join("..", "..", "hack", "crds"),
		},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = cmsv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = v1alpha2.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	webhookConfigSvc = initWebhookConfigService(webhookconfig.Config{CfgMapName: "test", CfgMapNamespace: "test"}, cfg)

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

func initWebhookConfigService(webhookCfg webhookconfig.Config, config *rest.Config) webhookconfig.AssetWebhookConfigService {
	dc, err := dynamic.NewForConfig(config)
	Expect(err).To(Succeed())

	configmapsResource := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	resourceGetter := dc.Resource(configmapsResource)
	webhookCfgService := webhookconfig.New(resourceGetter, webhookCfg.CfgMapName, webhookCfg.CfgMapNamespace)

	return webhookCfgService
}

func validateReconcilation(err error, result controllerruntime.Result) {
	Expect(err).ToNot(HaveOccurred())
	Expect(result.Requeue).To(BeFalse())
	Expect(result.RequeueAfter).To(Equal(60 * time.Hour))
}

func validateDocsTopic(status cmsv1alpha1.CommonDocsTopicStatus, meta controllerruntime.ObjectMeta, expectedPhase cmsv1alpha1.DocsTopicPhase, expectedReason cmsv1alpha1.DocsTopicReason) {
	Expect(status.Phase).To(Equal(expectedPhase))
	Expect(status.Reason).To(Equal(expectedReason))
}

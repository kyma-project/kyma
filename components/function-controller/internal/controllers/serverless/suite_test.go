package serverless

import (
	"context"
	"path/filepath"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/kubernetes"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/vrischmann/envconfig"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

var config FunctionConfig
var resourceClient resource.Client
var testEnv *envtest.Environment

const (
	testNamespace = "test-namespace-name"
)

func TestAPIs(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)

	ginkgo.RunSpecsWithDefaultAndCustomReporters(t,
		"Serverless Suite",
		[]ginkgo.Reporter{printer.NewlineReporter{}})
}

var _ = ginkgo.BeforeSuite(func(done ginkgo.Done) {
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(ginkgo.GinkgoWriter)))
	ginkgo.By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "..", "config", "crd", "bases"),
		},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(cfg).ToNot(gomega.BeNil())

	err = scheme.AddToScheme(scheme.Scheme)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	err = serverlessv1alpha1.AddToScheme(scheme.Scheme)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(k8sClient).ToNot(gomega.BeNil())

	resourceClient = resource.New(k8sClient, scheme.Scheme)

	err = envconfig.InitWithPrefix(&config, "TEST")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	rtm := serverlessv1alpha1.Nodejs12
	runtimeDockerfileConfigMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dockerfile-nodejs-12",
			Labels: map[string]string{kubernetes.ConfigLabel: "runtime",
				kubernetes.RuntimeLabel: string(rtm)},
			Namespace: testNamespace,
		},
	}
	ns := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	dockerRegistryConfiguration := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "serverless-registry-config-default",
			Namespace: testNamespace,
		},
		StringData: map[string]string{
			"registryAddress": "registry.kyma.local",
		},
	}
	gomega.Expect(resourceClient.Create(context.TODO(), &ns)).To(gomega.Succeed())
	gomega.Expect(resourceClient.Create(context.TODO(), &runtimeDockerfileConfigMap)).To(gomega.Succeed())
	gomega.Expect(resourceClient.Create(context.TODO(), &dockerRegistryConfiguration)).To(gomega.Succeed())

	close(done)
}, 60)

var _ = ginkgo.AfterSuite(func() {
	ginkgo.By("tearing down the test environment")
	err := testEnv.Stop()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
})

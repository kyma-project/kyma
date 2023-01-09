package kubernetes

import (
	"path/filepath"

	"github.com/onsi/gomega"
	"github.com/vrischmann/envconfig"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func setUpTestEnv(g *gomega.GomegaWithT) (cl client.Client, env *envtest.Environment) {
	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "..", "config", "crd", "bases"),
		},
		ErrorIfCRDPathMissing: true,
	}
	cfg, err := testEnv.Start()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(cfg).ToNot(gomega.BeNil())

	err = scheme.AddToScheme(scheme.Scheme)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(k8sClient).ToNot(gomega.BeNil())

	return k8sClient, testEnv
}

func tearDownTestEnv(g *gomega.GomegaWithT, testEnv *envtest.Environment) {
	g.Expect(testEnv.Stop()).To(gomega.Succeed())
}

func setUpControllerConfig(g *gomega.GomegaWithT) Config {
	var testCfg Config
	err := envconfig.InitWithPrefix(&testCfg, "TEST")
	g.Expect(err).To(gomega.BeNil())
	return testCfg
}

func compareConfigMaps(g *gomega.WithT, actual, expected *corev1.ConfigMap) {
	g.Expect(actual.GetLabels()).To(gomega.Equal(expected.GetLabels()))
	g.Expect(actual.GetAnnotations()).To(gomega.Equal(expected.GetAnnotations()))
	g.Expect(actual.Data).To(gomega.Equal(expected.Data))
	g.Expect(actual.BinaryData).To(gomega.Equal(expected.BinaryData))
}

func compareSecrets(g *gomega.WithT, actual, expected *corev1.Secret) {
	g.Expect(actual.GetLabels()).To(gomega.Equal(expected.GetLabels()))
	g.Expect(actual.GetAnnotations()).To(gomega.Equal(expected.GetAnnotations()))
	g.Expect(actual.Data).To(gomega.Equal(expected.Data))
}

func compareServiceAccounts(g *gomega.WithT, actual, expected *corev1.ServiceAccount) {
	g.Expect(actual.GetLabels()).To(gomega.Equal(expected.GetLabels()))
	g.Expect(actual.GetAnnotations()).To(gomega.Equal(expected.GetAnnotations()))
	g.Expect(actual.Secrets).To(gomega.Equal(expected.Secrets))
	g.Expect(actual.ImagePullSecrets).To(gomega.Equal(expected.ImagePullSecrets))
	g.Expect(actual.AutomountServiceAccountToken).To(gomega.Equal(expected.AutomountServiceAccountToken))
}

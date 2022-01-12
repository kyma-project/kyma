package serverless

import (
	"context"
	"fmt"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/kubernetes"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/onsi/gomega"
	"github.com/vrischmann/envconfig"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const (
	testNamespace = "test-namespace-name"
)

func setUpTestEnv(g *gomega.GomegaWithT) (cl resource.Client, env *envtest.Environment) {
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

	err = serverlessv1alpha1.AddToScheme(scheme.Scheme)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(k8sClient).ToNot(gomega.BeNil())

	resourceClient := resource.New(k8sClient, scheme.Scheme)
	g.Expect(resourceClient).ToNot(gomega.BeNil())
	return resourceClient, testEnv
}

func tearDownTestEnv(g *gomega.GomegaWithT, testEnv *envtest.Environment) {
	g.Expect(testEnv.Stop()).To(gomega.Succeed())
}

func setUpControllerConfig(g *gomega.GomegaWithT) FunctionConfig {
	var testCfg FunctionConfig
	err := envconfig.InitWithPrefix(&testCfg, "TEST")
	g.Expect(err).To(gomega.BeNil())
	return testCfg
}

func initializeServerlessResources(g *gomega.GomegaWithT, client resource.Client) {
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
	g.Expect(client.Create(context.TODO(), &ns)).To(gomega.Succeed())
	g.Expect(client.Create(context.TODO(), &dockerRegistryConfiguration)).To(gomega.Succeed())
}

func createDockerfileForRuntime(g *gomega.GomegaWithT, client resource.Client, rtm serverlessv1alpha1.Runtime) {
	runtimeDockerfileConfigMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("dockerfile-%s", string(rtm)),
			Labels: map[string]string{kubernetes.ConfigLabel: "runtime",
				kubernetes.RuntimeLabel: string(rtm)},
			Namespace: testNamespace,
		},
	}
	g.Expect(client.Create(context.TODO(), &runtimeDockerfileConfigMap)).To(gomega.Succeed())
}

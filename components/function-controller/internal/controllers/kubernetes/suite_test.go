package kubernetes

import (
	"context"
	"fmt"
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/vrischmann/envconfig"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
)

var (
	config            Config
	resourceClient    resource.Client
	k8sClient         client.Client
	testEnv           *envtest.Environment
	configMapSvc      ConfigMapService
	secretSvc         SecretService
	serviceAccountSvc ServiceAccountService
	roleSvc           RoleService
	roleBindingSvc    RoleBindingService
)

func TestAPIs(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)

	ginkgo.RunSpecsWithDefaultAndCustomReporters(t,
		"Kubernetes Suite",
		[]ginkgo.Reporter{printer.NewlineReporter{}})
}

var _ = ginkgo.BeforeSuite(func(done ginkgo.Done) {
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(ginkgo.GinkgoWriter)))
	ginkgo.By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(cfg).ToNot(gomega.BeNil())

	err = scheme.AddToScheme(scheme.Scheme)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(k8sClient).ToNot(gomega.BeNil())

	resourceClient = resource.New(k8sClient, scheme.Scheme)

	err = envconfig.InitWithPrefix(&config, "TEST")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	configMapSvc = NewConfigMapService(resourceClient, config)
	secretSvc = NewSecretService(resourceClient, config)
	serviceAccountSvc = NewServiceAccountService(resourceClient, config)
	roleSvc = NewRoleService(resourceClient, config)
	roleBindingSvc = NewRoleBindingService(resourceClient, config)

	baseNamespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: config.BaseNamespace}}
	gomega.Expect(resourceClient.Create(context.TODO(), baseNamespace)).To(gomega.Succeed())

	close(done)
}, 60)

var _ = ginkgo.AfterSuite(func() {
	ginkgo.By("tearing down the test environment")
	err := testEnv.Stop()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
})

func newFixBaseConfigMap(namespace, name string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", name),
			Namespace:    namespace,
			Labels:       map[string]string{ConfigLabel: RuntimeLabelValue},
		},
		Data:       map[string]string{"test_1": "value_!", "test_2": "value_2"},
		BinaryData: map[string][]byte{"test_1_b": []byte("value"), "test_2_b": []byte("value_2")},
	}
}

func newFixBaseSecret(namespace, name string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", name),
			Namespace:    namespace,
			Labels:       map[string]string{ConfigLabel: CredentialsLabelValue},
		},
		Data:       map[string][]byte{"key_1_b": []byte("value_1_b"), "key_2_b": []byte("value_2_b")},
		StringData: map[string]string{"key_1": "value_1", "key_2": "value_2"},
		Type:       "test",
	}
}

func newFixBaseSecretWithManagedLabel(namespace, name string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", name),
			Namespace:    namespace,
			Labels:       map[string]string{ConfigLabel: CredentialsLabelValue, ManagedByLabel: UserLabelValue},
		},
		Data:       map[string][]byte{"key_1_b": []byte("value_1_b"), "key_2_b": []byte("value_2_b")},
		StringData: map[string]string{"key_1": "value_1", "key_2": "value_2"},
		Type:       "test",
	}
}

func newFixBaseServiceAccount(namespace, name string) *corev1.ServiceAccount {
	falseValue := false
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", name),
			Namespace:    namespace,
			Labels:       map[string]string{ConfigLabel: ServiceAccountLabelValue},
		},
		Secrets:                      []corev1.ObjectReference{{Name: "test1"}, {Name: "test2"}},
		ImagePullSecrets:             []corev1.LocalObjectReference{{Name: "test-ips-1"}, {Name: "test-ips-2"}},
		AutomountServiceAccountToken: &falseValue,
	}
}

func newFixBaseRole(namespace, name string) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", name),
			Namespace:    namespace,
			Labels:       map[string]string{RbacLabel: RoleLabelValue},
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:         []string{"use"},
				APIGroups:     []string{"policy"},
				Resources:     []string{"podsecuritypolicies"},
				ResourceNames: []string{"serverless-build"},
			},
		},
	}
}

func newFixBaseRoleBinding(namespace, name, subjectNamespace string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", name),
			Namespace:    namespace,
			Labels:       map[string]string{RbacLabel: RoleBindingLabelValue},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "serverless",
				Namespace: subjectNamespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "serverless-build",
		},
	}
}

func newFixNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", name),
		},
	}
}

func compareConfigMaps(actual, expected *corev1.ConfigMap) {
	gomega.Expect(actual.GetLabels()).To(gomega.Equal(expected.GetLabels()))
	gomega.Expect(actual.GetAnnotations()).To(gomega.Equal(expected.GetAnnotations()))
	gomega.Expect(actual.Data).To(gomega.Equal(expected.Data))
	gomega.Expect(actual.BinaryData).To(gomega.Equal(expected.BinaryData))
}

func compareSecrets(actual, expected *corev1.Secret) {
	gomega.Expect(actual.GetLabels()).To(gomega.Equal(expected.GetLabels()))
	gomega.Expect(actual.GetAnnotations()).To(gomega.Equal(expected.GetAnnotations()))
	gomega.Expect(actual.Data).To(gomega.Equal(expected.Data))
}

func compareServiceAccounts(actual, expected *corev1.ServiceAccount) {
	gomega.Expect(actual.GetLabels()).To(gomega.Equal(expected.GetLabels()))
	gomega.Expect(actual.GetAnnotations()).To(gomega.Equal(expected.GetAnnotations()))
	gomega.Expect(actual.Secrets).To(gomega.Equal(expected.Secrets))
	gomega.Expect(actual.ImagePullSecrets).To(gomega.Equal(expected.ImagePullSecrets))
	gomega.Expect(actual.AutomountServiceAccountToken).To(gomega.Equal(expected.AutomountServiceAccountToken))
}

func compareRole(actual, expected *rbacv1.Role) {
	gomega.Expect(actual.GetLabels()).To(gomega.Equal(expected.GetLabels()))
	gomega.Expect(actual.GetAnnotations()).To(gomega.Equal(expected.GetAnnotations()))
	gomega.Expect(actual.Rules).To(gomega.Equal(expected.Rules))
}

func compareRoleBinding(actual, expected *rbacv1.RoleBinding) {
	gomega.Expect(actual.GetLabels()).To(gomega.Equal(expected.GetLabels()))
	gomega.Expect(actual.GetAnnotations()).To(gomega.Equal(expected.GetAnnotations()))
	gomega.Expect(actual.RoleRef).To(gomega.Equal(expected.RoleRef))
	gomega.Expect(actual.Subjects).To(gomega.Equal(expected.Subjects))
}

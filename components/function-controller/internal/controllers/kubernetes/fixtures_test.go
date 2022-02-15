package kubernetes

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{ConfigLabel: CredentialsLabelValue},
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
			Labels:       map[string]string{ConfigLabel: CredentialsLabelValue, v1alpha1.FunctionManagedByLabel: v1alpha1.FunctionResourceLabelUserValue},
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
			Name: name,
		},
	}
}

func createSecret(g *gomega.WithT, resourceClient resource.Client, secret *corev1.Secret) {
	g.Expect(resourceClient.Create(context.TODO(), secret)).To(gomega.Succeed())
}

func deleteSecret(g *gomega.WithT, k8sClient client.Client, secret *corev1.Secret) {
	g.Expect(k8sClient.Delete(context.TODO(), secret)).To(gomega.Succeed())
}

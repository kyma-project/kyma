package installer

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func CreateNamespace(k8sClient *corev1.CoreV1Client, name string) error {
	_, err := k8sClient.Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	})
	return err
}

func DeleteNamespace(k8sClient *corev1.CoreV1Client, name string) error {
	return k8sClient.Namespaces().Delete(name, nil)
}

package limit_range

import (
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ResourceQuotaClientInterface interface {
	CreateResourceQuota(namespace string, resourceQuota *v1.ResourceQuota) error
	DeleteResourceQuota(namespace string) error
}

type ResourceQuotaClient struct {
	Clientset *kubernetes.Clientset
}

func (lr *ResourceQuotaClient) CreateResourceQuota(namespace string, resourceQuota *v1.ResourceQuota) error {
	_, err := lr.Clientset.CoreV1().ResourceQuotas(namespace).Create(resourceQuota)
	if err != nil {
		return errors.Wrapf(err, "cannot create resource quota in namespace: %s", namespace)
	}
	return nil
}

func (lr *ResourceQuotaClient) DeleteResourceQuota(namespace string) error {
	err := lr.Clientset.CoreV1().ResourceQuotas(namespace).Delete(namespace, &metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrapf(err, "cannot delete resource quota from namespace: %s", namespace)
	}
	return nil
}

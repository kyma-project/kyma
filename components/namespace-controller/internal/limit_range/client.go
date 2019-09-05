package limit_range

import (
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type LimitRangesClientInterface interface {
	CreateLimitRange(namespace string, limitRange *v1.LimitRange) error
	DeleteLimitRange(namespace string) error
}

type LimitRangeClient struct {
	Clientset *kubernetes.Clientset
}

func (lr *LimitRangeClient) CreateLimitRange(namespace string, limitRange *v1.LimitRange) error {
	_, err := lr.Clientset.CoreV1().LimitRanges(namespace).Create(limitRange)
	if err != nil {
		return errors.Wrapf(err, "cannot create limit range in namespace: %s", namespace)
	}
	return nil
}

func (lr *LimitRangeClient) DeleteLimitRange(namespace string) error {
	err := lr.Clientset.CoreV1().LimitRanges(namespace).Delete(namespace, &metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrapf(err, "cannot delete limit range from namespace: %s", namespace)
	}
	return nil
}

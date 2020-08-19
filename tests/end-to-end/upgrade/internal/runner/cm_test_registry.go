package runner

import (
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

// ConfigMapTestRegistry allows to store information about passed tests in a config map
type ConfigMapTestRegistry struct {
	passed map[string]struct{}
	k8s    kubernetes.Interface

	cmName      string
	cmNamespace string
}

func NewConfigMapTestRegistry(k8s kubernetes.Interface, namespace, name string) (*ConfigMapTestRegistry, error) {
	cm, err := k8s.CoreV1().ConfigMaps(namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "while reading ConfigMap %s/%s with list of passed tests", namespace, name)
	}
	passed := make(map[string]struct{})
	for key := range cm.Data {
		passed[key] = struct{}{}
	}
	return &ConfigMapTestRegistry{
		passed:      passed,
		k8s:         k8s,
		cmName:      name,
		cmNamespace: namespace,
	}, nil
}

func (r *ConfigMapTestRegistry) IsTestPassed(name string) bool {
	_, found := r.passed[name]
	return found
}

func (r *ConfigMapTestRegistry) MarkTestPassed(name string) error {
	// Try to update, in case of error - retry
	// Every retry covers getting config map to avoid conflicts.
	return wait.ExponentialBackoff(retry.DefaultRetry, func() (bool, error) {
		cm, err := r.k8s.CoreV1().ConfigMaps(r.cmNamespace).Get(r.cmName, v1.GetOptions{})
		if err != nil {
			return false, nil
		}
		if cm.Data == nil {
			cm.Data = map[string]string{}
		}
		cm.Data[name] = "passed"
		_, err = r.k8s.CoreV1().ConfigMaps(r.cmNamespace).Update(cm)
		if err != nil {
			return false, nil
		}
		return true, nil
	})
}

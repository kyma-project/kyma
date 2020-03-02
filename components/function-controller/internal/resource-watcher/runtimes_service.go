package resource_watcher

import (
	"fmt"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	RuntimeLabelValue = "runtime"
)

type RuntimesService struct {
	coreClient *v1.CoreV1Client
	baseNamespace string
	cachedRuntimes map[string]*corev1.ConfigMap
}

func NewRuntimesService(coreClient *v1.CoreV1Client, baseNamespace string) *RuntimesService {
	return &RuntimesService{
		coreClient: coreClient,
		baseNamespace: baseNamespace,
		cachedRuntimes: nil,
	}
}

func (s *RuntimesService) GetRuntimes() (map[string]*corev1.ConfigMap, error) {
	if s.cachedRuntimes == nil || len(s.cachedRuntimes) == 0 {
		if err := s.UpdateCachedRuntimes(); err != nil {
			return nil, err
		}
	}
	return s.cachedRuntimes, nil
}

func (s *RuntimesService) UpdateCachedRuntimes() error {
	labelSelector := fmt.Sprintf("%s=%s", ConfigLabel, RuntimeLabelValue)
	list, err := s.coreClient.ConfigMaps(s.baseNamespace).List(metav1.ListOptions{
		LabelSelector:       labelSelector,
	})

	if err != nil {
		if apiErrors.IsNotFound(err) {
			return errors.Wrapf(err, "not found Runtimes in '%s' namespace by labelSelector '%s'", s.baseNamespace, labelSelector)
		}
		return errors.Wrapf(err, "while list Runtimes in '%s' namespace by labelSelector '%s'", s.baseNamespace, labelSelector)
	}
	if list == nil || len(list.Items) == 0 {
		return errors.New(fmt.Sprintf("not found Registry Credentials in '%s' namespace by labelSelector '%s'", s.baseNamespace, labelSelector))
	}

	if s.cachedRuntimes == nil {
		s.cachedRuntimes = make(map[string]*corev1.ConfigMap)
	}

	for _, runtime := range list.Items {
		key := fmt.Sprintf("%s/%s", s.baseNamespace, runtime.Name)
		s.cachedRuntimes[key] = &runtime
	}
	return nil
}

func (s *RuntimesService) UpdateCachedRuntime(namespace string) error {

}

func (s *RuntimesService) IsRuntime(configMap *corev1.ConfigMap) bool {
	hasRuntimeLabel := false
	for key, value := range configMap.GetLabels() {
		hasRuntimeLabel = key == ConfigLabel && value == RuntimeLabelValue
		if hasRuntimeLabel {
			break
		}
	}
	return hasRuntimeLabel
}

func (s *RuntimesService) IsBaseRuntime(configMap *corev1.ConfigMap) bool {
	return configMap.Namespace == s.baseNamespace && s.IsRuntime(configMap)
}

func (s *RuntimesService) ApplyRuntimesToNamespace(namespace string) error {
	runtimes, err := s.GetRuntimes()
	if err != nil {
		return errors.Wrapf(err, "while creating Runtimes in '%s' namespace", namespace)
	}

	for _, runtime := range runtimes {
		_, err = s.coreClient.ConfigMaps(namespace).Create(runtime)
		if err != nil {
			return errors.Wrapf(err, "while creating Runtime %v in '%s' namespace", *runtime, namespace)
		}
	}

	return nil
}

package resource_watcher

import (
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type RuntimesService struct {
	coreClient     *v1.CoreV1Client
	config         Config
	cachedRuntimes map[string]*corev1.ConfigMap
}

func NewRuntimesService(coreClient *v1.CoreV1Client, config Config) *RuntimesService {
	return &RuntimesService{
		coreClient:     coreClient,
		config:         config,
		cachedRuntimes: nil,
	}
}

func (s *RuntimesService) GetRuntimes() (map[string]*corev1.ConfigMap, error) {
	if s.cachedRuntimes == nil || len(s.cachedRuntimes) == 0 {
		if err := s.UpdateCachedRuntimes(); err != nil {
			return nil, errors.Wrap(err, "while getting Base Runtimes")
		}
	}
	return s.cachedRuntimes, nil
}

func (s *RuntimesService) GetRuntime(runtimeType string) (*corev1.ConfigMap, error) {
	runtimes, err := s.GetRuntimes()
	if err != nil {
		return nil, errors.Wrapf(err, "while getting '%s' runtime", runtimeType)
	}

	runtime := runtimes[runtimeType]
	if runtime == nil {
		return nil, errors.Wrapf(err, "while getting '%s' runtime - that runtime doesn't exists - check '%s' label", runtimeType, RuntimeLabel)
	}

	return runtimes[runtimeType], nil
}

func (s *RuntimesService) UpdateCachedRuntimes() error {
	labelSelector := fmt.Sprintf("%s=%s", ConfigLabel, RuntimeLabelValue)
	list, err := s.coreClient.ConfigMaps(s.config.BaseNamespace).List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})

	if err != nil {
		if apiErrors.IsNotFound(err) {
			return errors.Wrapf(err, "not found Runtimes in '%s' namespace by labelSelector '%s'", s.config.BaseNamespace, labelSelector)
		}
		return errors.Wrapf(err, "while list Runtimes in '%s' namespace by labelSelector '%s'", s.config.BaseNamespace, labelSelector)
	}
	if list == nil || len(list.Items) == 0 {
		return errors.New(fmt.Sprintf("not found Registry Credentials in '%s' namespace by labelSelector '%s'", s.config.BaseNamespace, labelSelector))
	}

	if s.cachedRuntimes == nil {
		s.cachedRuntimes = make(map[string]*corev1.ConfigMap)
	}

	for _, runtime := range list.Items {
		runtimeType := runtime.Labels[RuntimeLabel]
		if runtimeType != "" {
			s.cachedRuntimes[runtimeType] = &runtime
		}
	}
	return nil
}

func (s *RuntimesService) UpdateCachedRuntime(runtime *corev1.ConfigMap) error {
	if runtime == nil {
		return errors.New("runtime is null")
	}

	runtimeType := runtime.Labels[RuntimeLabel]
	if runtimeType == "" {
		return errors.New(fmt.Sprintf("runtime %v hasn't '%s' label", runtime, RuntimeLabel))
	}
	if s.cachedRuntimes == nil {
		s.cachedRuntimes = make(map[string]*corev1.ConfigMap)
	}
	s.cachedRuntimes[runtimeType] = runtime

	return nil
}

func (s *RuntimesService) CreateRuntimesInNamespace(namespace string) error {
	runtimes, err := s.GetRuntimes()
	if err != nil {
		return errors.Wrapf(err, "while creating Runtimes in '%s' namespace", namespace)
	}

	for _, runtime := range runtimes {
		newRuntime := s.copyRuntime(runtime, namespace)
		err := s.createRuntimeInNamespace(newRuntime, namespace)
		if err != nil {
			return errors.Wrapf(err, "while creating Runtimes in '%s' namespace", namespace)
		}
	}

	return nil
}

func (s *RuntimesService) UpdateRuntimesInNamespace(namespace string) error {
	runtimes, err := s.GetRuntimes()
	if err != nil {
		return errors.Wrapf(err, "while updating Runtimes in '%s' namespace", namespace)
	}

	for _, runtime := range runtimes {
		newRuntime := s.copyRuntime(runtime, namespace)
		err := s.updateRuntimeInNamespace(newRuntime, namespace)
		if err != nil {
			return errors.Wrapf(err, "while updating Runtimes in '%s' namespace", namespace)
		}
	}

	return nil
}

func (s *RuntimesService) UpdateRuntimeInNamespaces(runtime *corev1.ConfigMap, namespaces []string) error {
	for _, namespace := range namespaces {
		err := s.updateRuntimeInNamespace(runtime, namespace)
		if err != nil {
			return errors.Wrapf(err, "while updating Runtime '%s' in %v namespaces", runtime.Name, namespaces)
		}
	}
	return nil
}

func (s *RuntimesService) IsBaseRuntime(runtime *corev1.ConfigMap) bool {
	return runtime.Namespace == s.config.BaseNamespace && runtime.Labels[ConfigLabel] == RuntimeLabelValue
}

func (s *RuntimesService) createRuntimeInNamespace(runtime *corev1.ConfigMap, namespace string) error {
	_, err := s.coreClient.ConfigMaps(namespace).Create(runtime)
	if err != nil {
		if apiErrors.IsAlreadyExists(err) {
			return nil
		}
		return errors.Wrapf(err, "while creating Runtime '%s' in '%s' namespace", runtime.Name, namespace)
	}

	return nil
}

func (s *RuntimesService) updateRuntimeInNamespace(runtime *corev1.ConfigMap, namespace string) error {
	_, err := s.coreClient.ConfigMaps(namespace).Update(runtime)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			err = s.createRuntimeInNamespace(runtime, namespace)
			if err != nil {
				return err
			}
		} else {
			return errors.Wrapf(err, "while updating Runtime '%s' in '%s' namespace", runtime.Name, namespace)
		}
	}

	return nil
}

func (s *RuntimesService) copyRuntime(runtime *corev1.ConfigMap, namespace string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        runtime.Name,
			Namespace:   namespace,
			Labels:      runtime.Labels,
			Annotations: runtime.Annotations,
		},
		Data:       runtime.Data,
		BinaryData: runtime.BinaryData,
	}
}

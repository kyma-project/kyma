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
	RegistryCredentialsLabelValue = "registry-credentials"
)

type CredentialsService struct {
	coreClient        *v1.CoreV1Client
	baseNamespace     string
	cachedCredentials map[string]*corev1.Secret
}

func NewCredentialsService(coreClient *v1.CoreV1Client, baseNamespace string) *CredentialsService {
	return &CredentialsService{
		coreClient:        coreClient,
		baseNamespace:     baseNamespace,
		cachedCredentials: nil,
	}
}

func (s *CredentialsService) GetBaseCredentials() (*corev1.Secret, error) {
	return s.GetCredentialsFromNamespace(s.baseNamespace)
}

func (s *CredentialsService) GetCredentialsFromNamespace(namespace string) (*corev1.Secret, error) {
	credentials := s.cachedCredentials[namespace]
	if credentials == nil {
		if err := s.UpdateCachedCredentialsInNamespace(namespace); err != nil {
			return nil, err
		}
	}
	return credentials, nil
}

func (s *CredentialsService) UpdateCachedBaseCredentials() error {
	return s.UpdateCachedCredentialsInNamespace(s.baseNamespace)
}

func (s *CredentialsService) UpdateCachedCredentialsInNamespace(namespace string) error {
	labelSelector := fmt.Sprintf("%s=%s", ConfigLabel, RegistryCredentialsLabelValue)
	list, err := s.coreClient.Secrets(namespace).List(metav1.ListOptions{
		LabelSelector: labelSelector,
		Limit:         1,
	})

	if err != nil {
		if apiErrors.IsNotFound(err) {
			return errors.Wrapf(err, "not found Registry Credentials in '%s' namespace by labelSelector '%s'", namespace, labelSelector)
		}
		return errors.Wrapf(err, "while list Registry Credentials in '%s' namespace by labelSelector '%s'", namespace, labelSelector)
	}
	if list == nil || len(list.Items) == 0 {
		return errors.New(fmt.Sprintf("not found Registry Credentials in '%s' namespace by labelSelector '%s'", namespace, labelSelector))
	}

	s.cachedCredentials[namespace] = &list.Items[0]
	return nil
}

func (s *CredentialsService) IsCredentials(secret *corev1.Secret) bool {
	hasCredentialsLabel := false
	for key, value := range secret.GetLabels() {
		hasCredentialsLabel = key == ConfigLabel && value == RegistryCredentialsLabelValue
		if hasCredentialsLabel {
			break
		}
	}
	return hasCredentialsLabel
}

func (s *CredentialsService) IsBaseCredentials(secret *corev1.Secret) bool {
	return secret.Namespace == s.baseNamespace && s.IsCredentials(secret)
}

func (s *CredentialsService) CreateCredentialsInNamespace(namespace string) error {
	secret, err := s.GetBaseCredentials()
	if err != nil {
		return errors.Wrap(err, "while getting Base Registry Credentials")
	}

	_, err = s.coreClient.Secrets(namespace).Create(secret)
	if err != nil {
		return errors.Wrapf(err, "while creating Registry Credentials in '%s' namespace", namespace)
	}

	return nil
}

func (s *CredentialsService) UpdateCredentialsInNamespace(namespace string) error {
	secret, err := s.GetBaseCredentials()
	if err != nil {
		return errors.Wrap(err, "while getting Base Registry Credentials")
	}

	_, err = s.coreClient.Secrets(namespace).Update(secret)
	if err != nil {
		return errors.Wrapf(err, "while updating Registry Credentials in '%s' namespace", namespace)
	}

	return nil
}

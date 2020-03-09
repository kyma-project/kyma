package resource_watcher

import (
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type CredentialsService struct {
	coreClient        *v1.CoreV1Client
	config            Config
	cachedCredentials *corev1.Secret
}

func NewCredentialsService(coreClient *v1.CoreV1Client, config Config) *CredentialsService {
	return &CredentialsService{
		coreClient:        coreClient,
		config:            config,
		cachedCredentials: nil,
	}
}

func (s *CredentialsService) GetCredentials() (*corev1.Secret, error) {
	if s.cachedCredentials == nil {
		if err := s.UpdateCachedCredentials(nil); err != nil {
			return nil, errors.Wrap(err, "while getting Base Registry Credentials")
		}
	}
	return s.cachedCredentials, nil
}

func (s *CredentialsService) UpdateCachedCredentials(secret *corev1.Secret) error {
	if secret != nil {
		s.cachedCredentials = secret
		return nil
	}

	labelSelector := fmt.Sprintf("%s=%s", ConfigLabel, RegistryCredentialsLabelValue)
	list, err := s.coreClient.Secrets(s.config.BaseNamespace).List(metav1.ListOptions{
		LabelSelector: labelSelector,
		Limit:         1,
	})

	if err != nil {
		if apiErrors.IsNotFound(err) {
			return errors.Wrapf(err, "not found Base Registry Credentials in '%s' namespace by labelSelector '%s'", s.config.BaseNamespace, labelSelector)
		}
		return errors.Wrapf(err, "while list Base Registry Credentials in '%s' namespace by labelSelector '%s'", s.config.BaseNamespace, labelSelector)
	}
	if list == nil || len(list.Items) == 0 {
		return errors.New(fmt.Sprintf("not found Base Registry Credentials in '%s' namespace by labelSelector '%s'", s.config.BaseNamespace, labelSelector))
	}

	s.cachedCredentials = &list.Items[0]
	return nil
}

func (s *CredentialsService) CreateCredentialsInNamespace(namespace string) error {
	secret, err := s.GetCredentials()
	if err != nil {
		return errors.Wrapf(err, "while creating Registry Credentials in '%s' namespace", namespace)
	}
	newSecret := s.copyCredentials(secret, namespace)

	_, err = s.coreClient.Secrets(namespace).Create(newSecret)
	if err != nil {
		if apiErrors.IsAlreadyExists(err) {
			return nil
		}
		return errors.Wrapf(err, "while creating Registry Credentials in '%s' namespace", namespace)
	}

	return nil
}

func (s *CredentialsService) UpdateCredentialsInNamespace(namespace string) error {
	secret, err := s.GetCredentials()
	if err != nil {
		return errors.Wrapf(err, "while updating Registry Credentials in '%s' namespace", namespace)
	}
	newSecret := s.copyCredentials(secret, namespace)

	_, err = s.coreClient.Secrets(namespace).Update(newSecret)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			err = s.CreateCredentialsInNamespace(namespace)
			if err != nil {
				return err
			}
		} else {
			return errors.Wrapf(err, "while updating Registry Credentials in '%s' namespace", namespace)
		}
	}

	return nil
}

func (s *CredentialsService) UpdateCredentialsInNamespaces(namespaces []string) error {
	for _, namespace := range namespaces {
		err := s.UpdateCredentialsInNamespace(namespace)
		if err != nil {
			return errors.Wrapf(err, "while updating Registry Credentials in %v namespaces", namespaces)
		}
	}
	return nil
}

func (s *CredentialsService) IsBaseCredentials(secret *corev1.Secret) bool {
	return secret.Namespace == s.config.BaseNamespace && secret.Labels[ConfigLabel] == RegistryCredentialsLabelValue
}

func (s *CredentialsService) copyCredentials(secret *corev1.Secret, namespace string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        secret.Name,
			Namespace:   namespace,
			Labels:      secret.Labels,
			Annotations: secret.Annotations,
		},
		Data:       secret.Data,
		StringData: secret.StringData,
		Type:       secret.Type,
	}
}

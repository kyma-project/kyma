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
	cachedCredentials map[string]*corev1.Secret
	log               func(message string, args ...interface{})
}

func NewCredentialsService(coreClient *v1.CoreV1Client, config Config) *CredentialsService {
	return &CredentialsService{
		coreClient:        coreClient,
		config:            config,
		cachedCredentials: nil,
	}
}

func (s *CredentialsService) GetCredentials() (map[string]*corev1.Secret, error) {
	if s.cachedCredentials == nil || len(s.cachedCredentials) == 0 {
		if err := s.UpdateCachedCredentials(); err != nil {
			return nil, errors.Wrap(err, "while getting Credentials")
		}
	}
	return s.cachedCredentials, nil
}

func (s *CredentialsService) GetCredential(credentialType string) (*corev1.Secret, error) {
	credentials, err := s.GetCredentials()
	if err != nil {
		return nil, errors.Wrapf(err, "while getting '%s' Credential", credentialType)
	}

	runtime := credentials[credentialType]
	if runtime == nil {
		return nil, errors.Wrapf(err, "while getting '%s' Credential - that Credential doesn't exists - check '%s' label", credentialType, CredentialsLabel)
	}

	return credentials[credentialType], nil
}

func (s *CredentialsService) UpdateCachedCredentials() error {
	labelSelector := fmt.Sprintf("%s=%s", ConfigLabel, CredentialsLabelValue)
	list, err := s.coreClient.Secrets(s.config.BaseNamespace).List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})

	if err != nil {
		if apiErrors.IsNotFound(err) {
			return errors.Wrapf(err, "not found Base Credentials in '%s' namespace by labelSelector '%s'", s.config.BaseNamespace, labelSelector)
		}
		return errors.Wrapf(err, "while list Base Credentials in '%s' namespace by labelSelector '%s'", s.config.BaseNamespace, labelSelector)
	}
	if list == nil || len(list.Items) == 0 {
		return errors.New(fmt.Sprintf("not found Base Credentials in '%s' namespace by labelSelector '%s'", s.config.BaseNamespace, labelSelector))
	}

	if s.cachedCredentials == nil {
		s.cachedCredentials = make(map[string]*corev1.Secret)
	}

	s.log("\n\n%v\n\n", list)

	for _, credential := range list.Items {
		credentialType := credential.Labels[CredentialsLabel]
		if credentialType != "" {
			s.log("\n\n%s %v\n\n", credentialType, credential)
			s.cachedCredentials[credentialType] = &credential
		}
	}
	return nil
}

func (s *CredentialsService) UpdateCachedCredential(credential *corev1.Secret) error {
	if credential == nil {
		return errors.New("Credential is nil")
	}

	credentialType := credential.Labels[CredentialsLabel]
	if credentialType == "" {
		return errors.New(fmt.Sprintf("Credential %v hasn't '%s' label", credential, CredentialsLabel))
	}

	if s.cachedCredentials == nil {
		err := s.UpdateCachedCredentials()
		if err != nil {
			return nil
		}
	}

	s.cachedCredentials[credentialType] = credential
	return nil
}

func (s *CredentialsService) CreateCredentialsInNamespace(namespace string) error {
	credentials, err := s.GetCredentials()
	if err != nil {
		return errors.Wrapf(err, "while creating Runtimes in '%s' namespace", namespace)
	}

	for _, credential := range credentials {
		newCredential := s.copyCredentials(credential, namespace)
		err := s.createCredentialInNamespace(newCredential, namespace)
		if err != nil {
			return errors.Wrapf(err, "while creating Credentials in '%s' namespace", namespace)
		}
	}

	return nil
}

func (s *CredentialsService) UpdateCredentialsInNamespace(namespace string) error {
	credentials, err := s.GetCredentials()
	if err != nil {
		return errors.Wrapf(err, "while updating Runtimes in '%s' namespace", namespace)
	}

	for _, credential := range credentials {
		newCredential := s.copyCredentials(credential, namespace)
		err := s.updateCredentialInNamespace(newCredential, namespace)
		if err != nil {
			return errors.Wrapf(err, "while updating Credentials in '%s' namespace", namespace)
		}
	}

	return nil
}

func (s *CredentialsService) UpdateCredentialsInNamespaces(namespaces []string) error {
	for _, namespace := range namespaces {
		err := s.UpdateCredentialsInNamespace(namespace)
		if err != nil {
			return errors.Wrapf(err, "while updating Credentials in %v namespaces", namespaces)
		}
	}
	return nil
}

func (s *CredentialsService) IsBaseCredential(credential *corev1.Secret) bool {
	return credential.Namespace == s.config.BaseNamespace && credential.Labels[ConfigLabel] == CredentialsLabelValue
}

func (s *CredentialsService) SetLog(log func(message string, args ...interface{})) {
	s.log = log
}

func (s *CredentialsService) createCredentialInNamespace(credential *corev1.Secret, namespace string) error {
	_, err := s.coreClient.Secrets(namespace).Create(credential)
	if err != nil {
		if apiErrors.IsAlreadyExists(err) {
			return nil
		}
		return errors.Wrapf(err, "while creating Credential '%s' in '%s' namespace", credential.Name, namespace)
	}

	return nil
}

func (s *CredentialsService) updateCredentialInNamespace(credential *corev1.Secret, namespace string) error {
	_, err := s.coreClient.Secrets(namespace).Update(credential)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			err = s.createCredentialInNamespace(credential, namespace)
			if err != nil {
				return err
			}
		} else {
			return errors.Wrapf(err, "while updating Credential '%s' in '%s' namespace", credential.Name, namespace)
		}
	}

	return nil
}

func (s *CredentialsService) copyCredentials(credential *corev1.Secret, namespace string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        credential.Name,
			Namespace:   namespace,
			Labels:      credential.Labels,
			Annotations: credential.Annotations,
		},
		Data:       credential.Data,
		StringData: credential.StringData,
		Type:       credential.Type,
	}
}

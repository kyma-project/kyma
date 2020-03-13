package configwatcher

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type ServiceAccountService struct {
	coreClient           v1.CoreV1Interface
	config               Config
	cachedServiceAccount *corev1.ServiceAccount
}

func NewServiceAccountService(coreClient v1.CoreV1Interface, config Config) *ServiceAccountService {
	return &ServiceAccountService{
		coreClient:           coreClient,
		config:               config,
		cachedServiceAccount: nil,
	}
}

func (s *ServiceAccountService) GetServiceAccount() (*corev1.ServiceAccount, error) {
	if s.cachedServiceAccount == nil {
		if err := s.UpdateCachedServiceAccount(nil); err != nil {
			return nil, errors.Wrap(err, "while getting Base Service Account")
		}
	}
	return s.cachedServiceAccount, nil
}

func (s *ServiceAccountService) UpdateCachedServiceAccount(serviceAccount *corev1.ServiceAccount) error {
	if serviceAccount != nil {
		s.cachedServiceAccount = serviceAccount
		return nil
	}

	labelSelector := fmt.Sprintf("%s=%s", ConfigLabel, ServiceAccountLabelValue)
	list, err := s.coreClient.ServiceAccounts(s.config.BaseNamespace).List(metav1.ListOptions{
		LabelSelector: labelSelector,
		Limit:         1,
	})

	if err != nil {
		if apiErrors.IsNotFound(err) {
			return errors.Wrapf(err, "not found Base Service Account in '%s' namespace by labelSelector '%s'", s.config.BaseNamespace, labelSelector)
		}
		return errors.Wrapf(err, "while list Base Service Account in '%s' namespace by labelSelector '%s'", s.config.BaseNamespace, labelSelector)
	}
	if list == nil || len(list.Items) == 0 {
		return errors.New(fmt.Sprintf("not found Base Service Account in '%s' namespace by labelSelector '%s'", s.config.BaseNamespace, labelSelector))
	}

	s.cachedServiceAccount = &list.Items[0]
	return nil
}

func (s *ServiceAccountService) HandleServiceAccountInNamespace(namespace string) error {
	serviceAccount, err := s.GetServiceAccount()
	if err != nil {
		return errors.Wrapf(err, "while handling Service Account in '%s' namespace", namespace)
	}

	err = s.updateServiceAccountInNamespace(serviceAccount, namespace)
	if err != nil {
		return errors.Wrapf(err, "while handling Service Account in '%s' namespace", namespace)
	}

	return nil
}

func (s *ServiceAccountService) HandleServiceAccountInNamespaces(serviceAccount *corev1.ServiceAccount, namespaces []string) error {
	for _, namespace := range namespaces {
		err := s.updateServiceAccountInNamespace(serviceAccount, namespace)
		if err != nil {
			return errors.Wrapf(err, "while handling Service Account '%s' in %v namespaces", serviceAccount.Name, namespaces)
		}
	}
	return nil
}

func (s *ServiceAccountService) IsBaseServiceAccount(serviceAccount *corev1.ServiceAccount) bool {
	return serviceAccount.Namespace == s.config.BaseNamespace && serviceAccount.Labels[ConfigLabel] == ServiceAccountLabelValue
}

func (s *ServiceAccountService) createServiceAccountInNamespace(serviceAccount *corev1.ServiceAccount, namespace string) error {
	newServiceAccount := s.copyServiceAccount(serviceAccount, namespace)
	_, err := s.coreClient.ServiceAccounts(namespace).Create(newServiceAccount)
	if err != nil {
		if apiErrors.IsAlreadyExists(err) {
			return nil
		}
		return errors.Wrapf(err, "while creating Service Account '%s' in '%s' namespace", newServiceAccount.Name, namespace)
	}

	return nil
}

func (s *ServiceAccountService) updateServiceAccountInNamespace(serviceAccount *corev1.ServiceAccount, namespace string) error {
	newServiceAccount := s.copyServiceAccount(serviceAccount, namespace)
	_, err := s.coreClient.ServiceAccounts(namespace).Update(newServiceAccount)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			err = s.createServiceAccountInNamespace(serviceAccount, namespace)
			if err != nil {
				return err
			}
		} else {
			return errors.Wrapf(err, "while updating Service Account '%s' in '%s' namespace", newServiceAccount.Name, namespace)
		}
	}

	return nil
}

func (s *ServiceAccountService) copyServiceAccount(serviceAccount *corev1.ServiceAccount, namespace string) *corev1.ServiceAccount {
	secrets := s.shiftSecretTokens(serviceAccount)
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:        serviceAccount.Name,
			Namespace:   namespace,
			Labels:      serviceAccount.Labels,
			Annotations: serviceAccount.Annotations,
		},
		Secrets:          secrets,
		ImagePullSecrets: serviceAccount.ImagePullSecrets,
	}
}

func (s *ServiceAccountService) shiftSecretTokens(serviceAccount *corev1.ServiceAccount) []corev1.ObjectReference {
	secrets := make([]corev1.ObjectReference, 0)
	prefix := fmt.Sprintf("%s-token", serviceAccount.Name)
	for _, secret := range serviceAccount.Secrets {
		if !strings.HasPrefix(secret.Name, prefix) {
			secrets = append(secrets, secret)
		}
	}
	return secrets
}

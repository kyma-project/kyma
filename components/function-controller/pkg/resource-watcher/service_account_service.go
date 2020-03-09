package resource_watcher

import (
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type ServiceAccountService struct {
	coreClient           *v1.CoreV1Client
	config               Config
	cachedServiceAccount *corev1.ServiceAccount
	credentialsServices  *CredentialsService
}

func NewServiceAccountService(coreClient *v1.CoreV1Client, config Config, credentialsServices *CredentialsService) *ServiceAccountService {
	return &ServiceAccountService{
		coreClient:           coreClient,
		config:               config,
		cachedServiceAccount: nil,
		credentialsServices:  credentialsServices,
	}
}

func (s *ServiceAccountService) GetServiceAccount() (*corev1.ServiceAccount, error) {
	if s.cachedServiceAccount == nil {
		if err := s.UpdateCachedServiceAccount(); err != nil {
			return nil, errors.Wrap(err, "while getting Base Service Account")
		}
	}
	return s.cachedServiceAccount, nil
}

func (s *ServiceAccountService) UpdateCachedServiceAccount() error {
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

func (s *ServiceAccountService) CreateServiceAccountInNamespace(namespace string) error {
	serviceAccount, err := s.GetServiceAccount()
	if err != nil {
		return errors.Wrapf(err, "while creating Service Account in '%s' namespace", namespace)
	}

	newServiceAccount, err := s.copyServiceAccount(serviceAccount, namespace)
	if err != nil {
		return errors.Wrapf(err, "while creating Service Account in '%s' namespace", namespace)
	}

	_, err = s.coreClient.ServiceAccounts(namespace).Create(newServiceAccount)
	if err != nil {
		if apiErrors.IsAlreadyExists(err) {
			return nil
		}
		return errors.Wrapf(err, "while creating Service Account in '%s' namespace", namespace)
	}

	return nil
}

func (s *ServiceAccountService) copyServiceAccount(serviceAccount *corev1.ServiceAccount, namespace string) (*corev1.ServiceAccount, error) {
	secret, err := s.credentialsServices.GetCredential(RegistryCredentialsLabelValue)
	if err != nil {
		return nil, errors.Wrap(err, "while copying Service Account")
	}

	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:        serviceAccount.Name,
			Namespace:   namespace,
			Labels:      serviceAccount.Labels,
			Annotations: serviceAccount.Annotations,
		},
		Secrets: []corev1.ObjectReference{
			{
				Name: secret.Name,
			},
		},
	}, nil
}

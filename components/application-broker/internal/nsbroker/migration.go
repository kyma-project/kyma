package nsbroker

import (
	"fmt"
	"regexp"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scbeta "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedCorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/util/retry"
)

// MigrationService performs migration from old setup - one service per servicebroker to current solution
type MigrationService struct {
	serviceInterface   typedCorev1.ServicesGetter
	namespaceInterface typedCorev1.NamespaceInterface
	brokerGetter       scbeta.ServiceBrokersGetter

	svcServiceURLPattern *regexp.Regexp
	workingNs            string
	serviceName          string

	log logrus.FieldLogger
}

// NewMigrationService creates new MigrationService instance
func NewMigrationService(serviceInterface typedCorev1.ServicesGetter, brokerGetter scbeta.ServiceBrokersGetter, workingNamespace, serviceName string, log logrus.FieldLogger) (*MigrationService, error) {
	svcRegexp, err := regexp.Compile("http\\:\\/\\/([a-z][a-z0-9-]*)\\.")
	if err != nil {
		return nil, errors.Wrap(err, "while compiling regexp for URL of namespaced brokers")
	}

	return &MigrationService{
		svcServiceURLPattern: svcRegexp,
		serviceInterface:     serviceInterface,
		brokerGetter:         brokerGetter,
		log:                  log.WithField("service", "migration"),
		workingNs:            workingNamespace,
		serviceName:          serviceName,
	}, nil
}

// Migrate performs the migration
func (s *MigrationService) Migrate() {
	serviceBrokers, err := s.brokerGetter.ServiceBrokers(v1.NamespaceAll).List(v1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", BrokerLabelKey, BrokerLabelValue),
	})
	if err != nil {
		s.log.Error("Migration failed, could not list service brokers: %s", err.Error())
		return
	}

	for _, sb := range serviceBrokers.Items {
		err := s.migrateServiceBroker(&sb)
		if err != nil {
			s.log.Errorf("Migration has failed (%s/%s): %s", sb.Namespace, sb.Name, err.Error())
		}
	}
}

func (s *MigrationService) migrateServiceBroker(broker *v1beta1.ServiceBroker) error {
	// ensure the url is changed
	expectedURL := fmt.Sprintf("http://%s.%s.svc.cluster.local/%s", s.serviceName, s.workingNs, broker.Namespace)
	// check if the migration needs to be done
	if broker.Spec.URL == expectedURL {
		return nil
	}

	existingServiceName, err := s.getServiceNameFromBrokerURL(broker.Spec.URL)
	if err != nil {
		return errors.Wrapf(err, "while getting service name from URL in namespace %s", broker.Namespace)
	}

	s.log.Infof("Updating service broker %s/%s", broker.Namespace, broker.Name)
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		sb, err := s.brokerGetter.ServiceBrokers(broker.Namespace).Get(broker.Name, v1.GetOptions{})
		if err != nil {
			return err
		}
		sb.Spec.URL = expectedURL
		_, err = s.brokerGetter.ServiceBrokers(sb.Namespace).Update(sb)
		return err
	})
	if err != nil {
		return errors.Wrapf(err, "while updating ServiceBroker in namespace %s", broker.Namespace)
	}

	// do not delete the main service for application-broker
	if broker.Namespace == s.workingNs && existingServiceName == s.serviceName {
		return nil
	}
	// ensure the service is deleted
	s.log.Infof("Deleting service %s/%s", broker.Namespace, existingServiceName)
	err = s.serviceInterface.Services(broker.Namespace).Delete(existingServiceName, &v1.DeleteOptions{})
	switch {
	case err == nil:
	case apierrors.IsNotFound(err):
		return nil
	case err != nil:
		return errors.Wrapf(err, "while deleting Service %s", existingServiceName)
	}
	return nil
}

// getServiceNameFromBrokerURL extracts namespace from broker URL
func (s *MigrationService) getServiceNameFromBrokerURL(url string) (string, error) {
	out := s.svcServiceURLPattern.FindStringSubmatch(url)
	if len(out) < 2 {
		return "", fmt.Errorf("url:%s does not match pattern %s", url, s.svcServiceURLPattern.String())
	}
	// out[0] = matched regexp, out[1] = matched group in bracket
	return out[1], nil
}

package broker

import (
	"net/http"
	"time"

	mappingTypes "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	mappingCli "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	livenessInhibitor              = 10
	LivenessApplicationSampleName  = "informer.liveness.probe.application.name"
	livenessTestNamespace          = "kyma-system"
	applicationConnectorAPIVersion = "applicationconnector.kyma-project.io/v1alpha1"
)

type LivenessCheckStatus struct {
	Succeeded bool
}

// NewSanityChecker creates sanity checker service
func NewSanityChecker(mClient *mappingCli.Interface, log logrus.FieldLogger, livenessCheckStatus *LivenessCheckStatus) *SanityCheckService {
	return &SanityCheckService{
		mClient:             *mClient,
		log:                 log.WithField("service", "sanity checker"),
		counter:             0,
		livenessCheckStatus: livenessCheckStatus,
	}
}

// SanityCheckService performs sanity check for Application Broker
type SanityCheckService struct {
	mClient             mappingCli.Interface
	log                 logrus.FieldLogger
	counter             int
	livenessCheckStatus *LivenessCheckStatus
}

func (svc *SanityCheckService) SanityCheck() (int, error) {
	if svc.counter >= livenessInhibitor {
		svc.log.Info("Starting sanity check...")
		if err := svc.informerAvailability(); err != nil {
			svc.log.Errorf("failed to perform liveness check: %v ", err)
			return http.StatusInternalServerError, err
		}
		svc.counter = 0
		svc.log.Info("Finished sanity check")
	}
	svc.counter++

	return http.StatusOK, nil
}

func (svc *SanityCheckService) createSampleAppMapping() error {
	mapCli := svc.mClient.ApplicationconnectorV1alpha1().ApplicationMappings(livenessTestNamespace)
	mapping, err := mapCli.Create(&mappingTypes.ApplicationMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ApplicationMapping",
			APIVersion: applicationConnectorAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: LivenessApplicationSampleName,
		},
	})

	switch {
	case k8sErrors.IsAlreadyExists(err):
		svc.log.Errorf("sample Application Mapping already exists: %q", err)
		if err := svc.deleteSampleAppMapping(); err != nil {
			return errors.Wrapf(err, "while creating sample Application Mapping which already exists")
		}
		return nil
	case err != nil:
		return errors.Wrapf(err, "while creating sample Application Mapping")
	}

	svc.log.Infof("Mapping created, name [%s], namespace: [%s]", mapping.Name, mapping.Namespace)

	return nil
}

func (svc *SanityCheckService) deleteSampleAppMapping() error {
	return svc.mClient.ApplicationconnectorV1alpha1().ApplicationMappings(livenessTestNamespace).Delete(
		LivenessApplicationSampleName,
		&metav1.DeleteOptions{})
}

func (svc *SanityCheckService) deleteSamples() error {
	err := svc.deleteSampleAppMapping()
	if err != nil {
		return errors.Wrapf(err, "while deleting sample application mapping")
	}
	svc.log.Info("Deleted sample application mapping")
	return nil
}

func (svc *SanityCheckService) informerAvailability() error {

	defer func() {
		err := svc.deleteSamples()
		if err != nil {
			logrus.Errorf("while deleting sample resources: %v", err)
		}
	}()

	err := svc.createSampleAppMapping()
	if err != nil {
		return err
	}

	err = wait.Poll(1*time.Second, 5*time.Second, func() (done bool, err error) {
		if !svc.livenessCheckStatus.Succeeded {
			return false, errors.Errorf("liveness check failed - livenessCheckStatus.Succeeded flag equals %v", svc.livenessCheckStatus.Succeeded)
		}
		return true, nil
	})

	if err != nil {
		return err
	}

	svc.livenessCheckStatus.Succeeded = false
	return nil
}

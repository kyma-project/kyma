package broker

import (
	"net/http"
	"time"

	mappingTypes "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	mappingCli "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appOperator "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	appTypes "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	appCli "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	livenessInhibitor              = 10
	LivenessApplicationSampleName  = "informer.liveness.probe.application.name"
	LivenessAccessLabel            = "app-sample-liveness-label"
	LivenessProbeLabelKey          = "livenessAccessLabel"
	livenessTestNamespace          = "default"
	applicationConnectorAPIVersion = "applicationconnector.kyma-project.io/v1alpha1"
)

// NewSanityChecker creates sanity checker service
func NewSanityChecker(appClient *appCli.Interface, mClient *mappingCli.Interface, k8sClient kubernetes.Interface, log logrus.FieldLogger) *SanityCheckService {
	return &SanityCheckService{
		appClient: *appClient,
		mClient:   *mClient,
		k8sClient: k8sClient,
		log:       log.WithField("service", "sanity checker"),
		counter:   0,
	}
}

// SanityCheckService performs sanity check for Application Broker
type SanityCheckService struct {
	appClient appCli.Interface
	mClient   mappingCli.Interface
	k8sClient kubernetes.Interface
	log       logrus.FieldLogger
	counter   int
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

func (svc *SanityCheckService) createSampleApp() error {
	_, err := svc.appClient.ApplicationconnectorV1alpha1().Applications().Create(
		&appOperator.Application{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Application",
				APIVersion: applicationConnectorAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: LivenessApplicationSampleName,
			},
			Spec: appTypes.ApplicationSpec{
				AccessLabel: LivenessAccessLabel,
				Services:    []appTypes.Service{},
			},
		},
	)

	switch {
	case k8sErrors.IsAlreadyExists(err):
		svc.log.Errorf("sample Application already exists: %q", err)
		if err := svc.deleteSampleApp(); err != nil {
			return errors.Wrapf(err, "while creating sample Application which already exists")
		}
		return nil
	case err != nil:
		return errors.Wrap(err, "while creating sample Application")
	}

	svc.log.Info("created sample Application")
	return nil
}

func (svc *SanityCheckService) deleteSampleApp() error {
	return svc.appClient.ApplicationconnectorV1alpha1().Applications().Delete(
		LivenessApplicationSampleName,
		&metav1.DeleteOptions{})
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
	err := svc.deleteSampleApp()
	if err != nil {
		return errors.Wrapf(err, "while deleting sample application")
	}
	err = svc.deleteSampleAppMapping()
	if err != nil {
		return errors.Wrapf(err, "while deleting sample application mapping")
	}
	svc.log.Info("Deleted sample resources: application, application mapping")
	return nil
}

func (svc *SanityCheckService) informerAvailability() error {

	defer func() {
		err := svc.deleteSamples()
		if err != nil {
			logrus.Errorf("while deleting sample resources: %v", err)
		}
	}()

	err := svc.createSampleApp()
	if err != nil {
		return err
	}

	err = svc.createSampleAppMapping()
	if err != nil {
		return err
	}

	err = wait.Poll(1*time.Second, 30*time.Second, func() (done bool, err error) {
		resNs, err := svc.k8sClient.CoreV1().Namespaces().Get(
			livenessTestNamespace,
			metav1.GetOptions{})
		if err != nil {
			svc.log.Errorf("while getting namespace %s", err)
			return false, nil
		}

		if resNs.Labels[LivenessProbeLabelKey] != LivenessAccessLabel {
			return false, errors.New("sample access label not matching")
		}

		return true, nil
	})

	if err != nil {
		return err
	}

	return nil
}

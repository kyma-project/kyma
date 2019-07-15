package testsuite

import (
	"github.com/avast/retry-go"
	acApi "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	acClient "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/consts"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/step"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CreateApplication struct {
	applications     acClient.ApplicationsGetter
	skipInstallation bool
}

var _ step.Step = &CreateApplication{}

func NewCreateApplication(applications acClient.ApplicationsGetter, skipInstallation bool) *CreateApplication {
	return &CreateApplication{
		applications:     applications,
		skipInstallation: skipInstallation,
	}
}

func (s *CreateApplication) Name() string {
	return "Create application"
}

func (s *CreateApplication) Run() error {
	spec := acApi.ApplicationSpec{
		Services:         []acApi.Service{},
		AccessLabel:      consts.AccessLabel,
		SkipInstallation: s.skipInstallation,
	}

	dummyApp := &acApi.Application{
		TypeMeta:   v1.TypeMeta{Kind: "Application", APIVersion: acApi.SchemeGroupVersion.String()},
		ObjectMeta: v1.ObjectMeta{Name: consts.AppName},
		Spec:       spec,
	}

	_, err := s.applications.Applications().Create(dummyApp)
	if err != nil {
		return err
	}

	return retry.Do(s.isApplicationReady)
}

func (s *CreateApplication) isApplicationReady() error {
	application, err := s.applications.Applications().Get(consts.AppName, v1.GetOptions{})

	if err != nil {
		return err
	}

	if application.Status.InstallationStatus.Status == "DEPLOYED" {
		return errors.New("Unexpected installation status: " + application.Status.InstallationStatus.Status)
	}

	return nil
}

func (s *CreateApplication) Cleanup() error {
	return s.applications.Applications().Delete(consts.AppName, &v1.DeleteOptions{})
}

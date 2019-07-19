package testsuite

import (
	"fmt"
	"github.com/avast/retry-go"
	acApi "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	acClient "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateApplication is a step which creates new Application
type CreateApplication struct {
	applications     acClient.ApplicationInterface
	skipInstallation bool
	name             string
	accessLabel      string
}

var _ step.Step = &CreateApplication{}

// NewCreateApplication returns new CreateApplication
func NewCreateApplication(name, accessLabel string, skipInstallation bool, applications acClient.ApplicationInterface, ) *CreateApplication {
	return &CreateApplication{
		name:             name,
		applications:     applications,
		skipInstallation: skipInstallation,
		accessLabel:      accessLabel,
	}
}

// Name returns name name of the step
func (s *CreateApplication) Name() string {
	return fmt.Sprintf("Create application %s", s.name)
}

// Run executes the step
func (s *CreateApplication) Run() error {
	spec := acApi.ApplicationSpec{
		Services:         []acApi.Service{},
		AccessLabel:      s.accessLabel,
		SkipInstallation: s.skipInstallation,
	}

	dummyApp := &acApi.Application{
		TypeMeta:   v1.TypeMeta{Kind: "Application", APIVersion: acApi.SchemeGroupVersion.String()},
		ObjectMeta: v1.ObjectMeta{Name: s.name},
		Spec:       spec,
	}

	_, err := s.applications.Create(dummyApp)
	if err != nil {
		return err
	}

	return retry.Do(s.isApplicationReady)
}

func (s *CreateApplication) isApplicationReady() error {
	application, err := s.applications.Get(s.name, v1.GetOptions{})
	if err != nil {
		return err
	}

	if application.Status.InstallationStatus.Status == "DEPLOYED" {
		return errors.Errorf("unexpected installation status: %s", application.Status.InstallationStatus.Status)
	}

	return nil
}

// Cleanup removes all resources that may possibly created by the step
func (s *CreateApplication) Cleanup() error {
	err := s.applications.Delete(s.name, &v1.DeleteOptions{})
	if err != nil {
		return err
	}

	return helpers.AwaitResourceDeleted(func() (interface{}, error) {
		return s.applications.Get(s.name, v1.GetOptions{})
	})
}

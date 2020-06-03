package testsuite

import (
	"fmt"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appconnectorv1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	appconnectorclientset "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	esclientset "github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset/typed/sources/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
)

// CreateApplication is a step which creates new Application
type CreateApplication struct {
	applications     appconnectorclientset.ApplicationInterface
	httpSources      esclientset.HTTPSourceInterface
	skipInstallation bool
	name             string
	accessLabel      string
	tenant           string
	group            string
}

var _ step.Step = &CreateApplication{}

// NewCreateApplication returns new CreateApplication
func NewCreateApplication(name, accessLabel string, skipInstallation bool, tenant, group string,
	applications appconnectorclientset.ApplicationInterface, httpSourceClient esclientset.HTTPSourceInterface) *CreateApplication {
	return &CreateApplication{
		name:             name,
		applications:     applications,
		httpSources:      httpSourceClient,
		skipInstallation: skipInstallation,
		accessLabel:      accessLabel,
		tenant:           tenant,
		group:            group,
	}
}

// Name returns name name of the step
func (s *CreateApplication) Name() string {
	return fmt.Sprintf("Create application %s", s.name)
}

// Run executes the step
func (s *CreateApplication) Run() error {
	spec := appconnectorv1alpha1.ApplicationSpec{
		Services:         []appconnectorv1alpha1.Service{},
		AccessLabel:      s.accessLabel,
		SkipInstallation: s.skipInstallation,
		Tenant:           s.tenant,
		Group:            s.group,
	}

	dummyApp := &appconnectorv1alpha1.Application{
		TypeMeta:   metav1.TypeMeta{Kind: "Application", APIVersion: appconnectorv1alpha1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: s.name},
		Spec:       spec,
	}

	_, err := s.applications.Create(dummyApp)
	if err != nil {
		return err
	}

	return retry.Do(s.isApplicationReady)
}

func (s *CreateApplication) isApplicationReady() error {
	application, err := s.applications.Get(s.name, metav1.GetOptions{})

	if err != nil {
		return err
	}

	if application.Status.InstallationStatus.Status == "DEPLOYED" {
		return errors.Errorf("unexpected installation status: %s", application.Status.InstallationStatus.Status)
	}

	return retry.Do(s.isHttpSourceReady)
}

func (s *CreateApplication) isHttpSourceReady() error {
	if s.httpSources != nil {
		httpsource, err := s.httpSources.Get(s.name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if !httpsource.Status.IsReady() {
			return errors.Errorf("httpsource %s is not ready. Status of HttpSource: \n %+v", httpsource.Name, httpsource.Status)
		}
	}
	return nil
}

// Cleanup removes all resources that may possibly created by the step
func (s *CreateApplication) Cleanup() error {
	err := s.applications.Delete(s.name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return helpers.AwaitResourceDeleted(func() (interface{}, error) {
		return s.applications.Get(s.name, metav1.GetOptions{})
	})
}

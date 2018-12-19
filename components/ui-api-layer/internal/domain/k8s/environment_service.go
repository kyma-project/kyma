package k8s

import (
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/application/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const envLabelSelector = "env=true"

type environmentService struct {
	appLister   ApplicationLister
	nsInterface corev1.NamespaceInterface
}

func newEnvironmentService(nsInterface corev1.NamespaceInterface, appLister ApplicationLister) *environmentService {
	return &environmentService{
		nsInterface: nsInterface,
		appLister:   appLister,
	}
}

func (svc *environmentService) List() ([]gqlschema.Environment, error) {
	list, err := svc.nsInterface.List(metav1.ListOptions{
		LabelSelector: envLabelSelector, // namespaces with label env=true are treated as environments
	})
	if err != nil {
		return []gqlschema.Environment{}, errors.Wrapf(err, "while listing %s", pretty.ApplicationMapping)
	}

	result := make([]gqlschema.Environment, 0)
	for _, ns := range list.Items {
		items, err := svc.appLister.ListInEnvironment(ns.Name)
		if err != nil {
			return []gqlschema.Environment{}, errors.Wrapf(err, "while listing %s for env", pretty.Application)
		}
		appNames := make([]string, 0)
		for _, app := range items {
			appNames = append(appNames, app.Name)
		}

		result = append(result, gqlschema.Environment{
			Name:         ns.Name,
			Applications: appNames,
		})
	}

	return result, nil
}

func (svc *environmentService) ListForApplication(appName string) ([]gqlschema.Environment, error) {
	namespaces, err := svc.appLister.ListNamespacesFor(appName)
	if err != nil {
		return []gqlschema.Environment{}, errors.Wrap(err, "while listing namespaces")
	}

	result := make([]gqlschema.Environment, 0)
	for _, ns := range namespaces {
		items, err := svc.appLister.ListInEnvironment(ns)
		if err != nil {
			return []gqlschema.Environment{}, errors.Wrapf(err, "while listing %s", pretty.Application)
		}
		appNames := make([]string, 0)
		for _, app := range items {
			appNames = append(appNames, app.Name)
		}
		result = append(result, gqlschema.Environment{
			Name:         ns,
			Applications: appNames,
		})
	}
	return result, nil
}

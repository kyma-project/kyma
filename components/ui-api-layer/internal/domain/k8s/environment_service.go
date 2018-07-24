package k8s

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
)

const envLabelSelector = "env=true"

type environmentService struct {
	reLister    RemoteEnvironmentLister
	nsInterface corev1.NamespaceInterface
}

func newEnvironmentService(nsInterface corev1.NamespaceInterface, reLister RemoteEnvironmentLister) *environmentService {
	return &environmentService{
		nsInterface: nsInterface,
		reLister:    reLister,
	}
}

func (svc *environmentService) List() ([]gqlschema.Environment, error) {
	list, err := svc.nsInterface.List(metav1.ListOptions{
		LabelSelector: envLabelSelector, // namespaces with label env=true are treated as environments
	})
	if err != nil {
		return []gqlschema.Environment{}, errors.Wrap(err, "while listing environment mappings")
	}

	result := make([]gqlschema.Environment, 0)
	for _, ns := range list.Items {
		res, err := svc.reLister.ListInEnvironment(ns.Name)
		if err != nil {
			return []gqlschema.Environment{}, errors.Wrap(err, "while listing remote envs for env")
		}
		reNames := make([]string, 0)
		for _, re := range res {
			reNames = append(reNames, re.Name)
		}

		result = append(result, gqlschema.Environment{
			Name:               ns.Name,
			RemoteEnvironments: reNames,
		})
	}

	return result, nil
}

func (svc *environmentService) ListForRemoteEnvironment(reName string) ([]gqlschema.Environment, error) {
	namespaces, err := svc.reLister.ListNamespacesFor(reName)
	if err != nil {
		return []gqlschema.Environment{}, errors.Wrap(err, "while listing namespaces")
	}

	result := make([]gqlschema.Environment, 0)
	for _, ns := range namespaces {
		res, err := svc.reLister.ListInEnvironment(ns)
		if err != nil {
			return []gqlschema.Environment{}, errors.Wrap(err, "while listing remote envs")
		}
		remoteEnvNames := make([]string, 0)
		for _, re := range res {
			remoteEnvNames = append(remoteEnvNames, re.Name)
		}
		result = append(result, gqlschema.Environment{
			Name:               ns,
			RemoteEnvironments: remoteEnvNames,
		})
	}
	return result, nil
}

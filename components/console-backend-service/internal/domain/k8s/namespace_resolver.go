package k8s

import (
	"context"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"

	"github.com/golang/glog"
	appPretty "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

//go:generate mockery -name=namespaceSvc -output=automock -outpkg=automock -case=underscore
type namespaceSvc interface {
	Create(name string, labels gqlschema.Labels) (*v1.Namespace, error)
	Update(name string, labels gqlschema.Labels) (*v1.Namespace, error)
	List() ([]*v1.Namespace, error)
	Find(name string) (*v1.Namespace, error)
	Delete(name string) error
}

//go:generate mockery -name=gqlNamespaceConverter -output=automock -outpkg=automock -case=underscore
type gqlNamespaceConverter interface {
	ToGQLs(in []*v1.Namespace) ([]gqlschema.Namespace, error)
	ToGQL(in *v1.Namespace) (*gqlschema.Namespace, error)
	ToGQLsWithPods(in []NamespaceWithAdditionalData) ([]gqlschema.Namespace, error)
	ToGQLWithPods(in NamespaceWithAdditionalData) (*gqlschema.Namespace, error)
}

type namespaceResolver struct {
	namespaceSvc       namespaceSvc
	podSvc	           podSvc
	appRetriever       shared.ApplicationRetriever
	namespaceConverter gqlNamespaceConverter
	systemNamespaces   []string
}

func newNamespaceResolver(namespaceSvc namespaceSvc, podSvc podSvc, appRetriever shared.ApplicationRetriever, systemNamespaces []string) *namespaceResolver {
	return &namespaceResolver{
		namespaceSvc:       namespaceSvc,
		podSvc:             podSvc,
		appRetriever:       appRetriever,
		systemNamespaces:   systemNamespaces,
		namespaceConverter: &namespaceConverter{},
	}
}

type NamespaceWithAdditionalData struct {
	namespace *v1.Namespace
	isSystemNamespace bool
	pods []*v1.Pod
}

func (r *namespaceResolver) NamespacesQuery(ctx context.Context, withSystemNamespaces *bool) ([]gqlschema.Namespace, error) {
	var err error

	var namespaces []NamespaceWithAdditionalData
	rawNamespaces, err := r.namespaceSvc.List()

	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}

		glog.Error(errors.Wrapf(err, "while listing %s", pretty.Namespaces))
		return nil, gqlerror.New(err, pretty.Namespaces)
	}

	for _, ns := range rawNamespaces {
		pods, err := r.podSvc.List(ns.Name, pager.PagingParams{
			First:  nil,
			Offset: nil,
		})
		if err != nil {
			glog.Error(errors.Wrapf(err, "while getting the list of pods for %s %s", ns.Name, pretty.Namespace))
		}

		isSystem := isSystem(*ns, r.systemNamespaces)
		if !isSystem || (withSystemNamespaces != nil && *withSystemNamespaces && isSystem) {
			namespaces = append(namespaces, NamespaceWithAdditionalData{
				namespace: ns,
				isSystemNamespace: isSystem,
				pods: pods,
			})
		}
	}

	converted, err := r.namespaceConverter.ToGQLsWithPods(namespaces)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.Namespaces))
		return nil, gqlerror.New(err, pretty.Namespaces)
	}

	return converted, nil
}

func (r *namespaceResolver) ApplicationsField(ctx context.Context, obj *gqlschema.Namespace) ([]string, error) {
	if obj == nil {
		return nil, errors.New("Cannot get application field for namespace")
	}

	items, err := r.appRetriever.Application().ListInNamespace(obj.Name)
	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}

		return nil, errors.Wrapf(err, "while listing %s for namespace %s", appPretty.Application, obj.Name)
	}

	var appNames []string
	for _, app := range items {
		appNames = append(appNames, app.Name)
	}

	return appNames, nil
}

func (r *namespaceResolver) NamespaceQuery(ctx context.Context, name string) (*gqlschema.Namespace, error) {
	namespace, err := r.namespaceSvc.Find(name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s with name %s", pretty.Namespace, name))
		return nil, gqlerror.New(err, pretty.Namespace, gqlerror.WithName(name))
	}

	pods, err := r.podSvc.List(namespace.Name, pager.PagingParams{
		First:  nil,
		Offset: nil,
	})
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting the list of pods for %s %s", namespace.Name, pretty.Namespace))
	}

	isSystem := isSystem(*namespace, r.systemNamespaces)

	ns := NamespaceWithAdditionalData{
		namespace: namespace,
		isSystemNamespace: isSystem,
		pods: pods,
	}

	converted, err := r.namespaceConverter.ToGQLWithPods(ns)

	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.Namespace))
		return nil, gqlerror.New(err, pretty.Namespace, gqlerror.WithName(name))
	}

	return converted, nil
}

func (r *namespaceResolver) CreateNamespace(ctx context.Context, name string, labels *gqlschema.Labels) (gqlschema.NamespaceMutationOutput, error) {
	gqlLabels := r.populateLabels(labels)
	ns, err := r.namespaceSvc.Create(name, gqlLabels)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s `%s`", pretty.Namespace, name))
		return gqlschema.NamespaceMutationOutput{}, gqlerror.New(err, pretty.Namespace, gqlerror.WithName(name))
	}
	return gqlschema.NamespaceMutationOutput{
		Name:   name,
		Labels: ns.Labels,
	}, nil
}

func (r *namespaceResolver) UpdateNamespace(ctx context.Context, name string, labels gqlschema.Labels) (gqlschema.NamespaceMutationOutput, error) {
	gqlLabels := r.populateLabels(&labels)
	ns, err := r.namespaceSvc.Update(name, gqlLabels)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while editing %s `%s`", pretty.Namespace, name))
		return gqlschema.NamespaceMutationOutput{}, gqlerror.New(err, pretty.Namespace, gqlerror.WithName(name))
	}
	return gqlschema.NamespaceMutationOutput{
		Name:   name,
		Labels: ns.Labels,
	}, nil
}

func (r *namespaceResolver) DeleteNamespace(ctx context.Context, name string) (*gqlschema.Namespace, error) {
	namespace, err := r.namespaceSvc.Find(name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s with name %s", pretty.Namespace, name))
		return nil, gqlerror.New(err, pretty.Namespace, gqlerror.WithName(name))
	}

	namespaceCopy := namespace.DeepCopy()
	deletedNamespace, err := r.namespaceConverter.ToGQL(namespaceCopy)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.Namespace))
		return nil, gqlerror.New(err, pretty.Namespace, gqlerror.WithName(name))
	}

	err = r.namespaceSvc.Delete(name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s `%s`", pretty.Namespace, name))
		return nil, gqlerror.New(err, pretty.Namespace, gqlerror.WithName(name))
	}

	return deletedNamespace, nil
}

func (r *namespaceResolver) populateLabels(givenLabels *gqlschema.Labels) map[string]string {
	labels := map[string]string{}
	if givenLabels != nil {
		for k, v := range *givenLabels {
			labels[k] = v
		}
	}
	return labels
}

func isSystem(namespace v1.Namespace, sysNamespaces []string) bool {
	for _, sysNs := range sysNamespaces {
		if sysNs == namespace.Name {
			return true
		}
	}
	return false
}
package k8s

import (
	"context"

	"github.com/golang/glog"
	appPretty "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
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
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

type namespaceResolver struct {
	namespaceSvc       namespaceSvc
	appRetriever       shared.ApplicationRetriever
	namespaceConverter namespaceConverter
	systemNamespaces   []string
}

func newNamespaceResolver(namespaceSvc namespaceSvc, appRetriever shared.ApplicationRetriever, systemNamespaces []string) *namespaceResolver {
	return &namespaceResolver{
		namespaceSvc:       namespaceSvc,
		appRetriever:       appRetriever,
		namespaceConverter: *newNamespaceConverter(systemNamespaces),
		systemNamespaces:   systemNamespaces,
	}
}

func (r *namespaceResolver) NamespacesQuery(ctx context.Context, withSystemNamespaces *bool, withInactiveStatus *bool) ([]gqlschema.Namespace, error) {
	namespaces, err := r.namespaceSvc.List()

	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}

		glog.Error(errors.Wrapf(err, "while listing %s", pretty.Namespaces))
		return nil, gqlerror.New(err, pretty.Namespaces)
	}

	var filteredNamespaces []*v1.Namespace
	for _, ns := range namespaces {
		if r.checkNamespace(ns, withSystemNamespaces, withInactiveStatus) {
			filteredNamespaces = append(filteredNamespaces, ns)
		}
	}

	converted := r.namespaceConverter.ToGQLs(filteredNamespaces)

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

	converted := r.namespaceConverter.ToGQL(namespace)

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
	deletedNamespace := r.namespaceConverter.ToGQL(namespaceCopy)

	err = r.namespaceSvc.Delete(name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s `%s`", pretty.Namespace, name))
		return nil, gqlerror.New(err, pretty.Namespace, gqlerror.WithName(name))
	}

	return deletedNamespace, nil
}

func (r *namespaceResolver) NamespaceEventSubscription(ctx context.Context, withSystemNamespaces *bool) (<-chan gqlschema.NamespaceEvent, error) {
	namespaceChannel := make(chan gqlschema.NamespaceEvent, 1)
	filter := func(namespace *v1.Namespace) bool {
		newBool := true
		return namespace != nil && r.checkNamespace(namespace, withSystemNamespaces, &newBool)
	}
	namespaceListener := listener.NewNamespace(namespaceChannel, filter, &r.namespaceConverter, r.systemNamespaces)

	r.namespaceSvc.Subscribe(namespaceListener)

	go func() {
		defer close(namespaceChannel)
		defer r.namespaceSvc.Unsubscribe(namespaceListener)
		<-ctx.Done()
	}()

	return namespaceChannel, nil
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

func (r *namespaceResolver) checkNamespace(ns *v1.Namespace, withSystemNamespaces *bool, withInactiveStatus *bool) bool {
	isSystem := isSystemNamespace(*ns, r.systemNamespaces)
	passedSystemNamespaceCheck := !isSystem || (withSystemNamespaces != nil && *withSystemNamespaces && isSystem)
	passedStatusCheck := ns.Status.Phase == "Active" || (withInactiveStatus != nil && *withInactiveStatus)
	return passedSystemNamespaceCheck && passedStatusCheck
}

package k8s

import (
	"context"

	"github.com/golang/glog"
	appPretty "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: Write tests

//go:generate mockery -name=namespaceSvc -output=automock -outpkg=automock -case=underscore
type namespaceSvc interface {
	Create(name string, labels gqlschema.Labels) (*v1.Namespace, error)
	List() ([]*v1.Namespace, error)
}

type namespaceResolver struct {
	namespaceSvc namespaceSvc
	appRetriever shared.ApplicationRetriever
}

func newNamespaceResolver(namespaceSvc namespaceSvc, appRetriever shared.ApplicationRetriever) *namespaceResolver {
	return &namespaceResolver{
		namespaceSvc: namespaceSvc,
		appRetriever: appRetriever,
	}
}

// TODO: Split this query into two
func (r *namespaceResolver) NamespacesQuery(ctx context.Context, applicationName *string) ([]gqlschema.Namespace, error) {
	var err error

	var namespaces []*v1.Namespace
	if applicationName == nil {
		namespaces, err = r.namespaceSvc.List()
	} else {

		// TODO: Investigate if we still need the query for namespaces bound to an application
		var namespaceNames []string
		namespaceNames, err = r.appRetriever.Application().ListNamespacesFor(*applicationName)

		//TODO: Refactor 'ListNamespacesFor' to return []v1.Namespace and remove this workaround
		for _, nsName := range namespaceNames {
			namespaces = append(namespaces, &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: nsName,
				},
			})
		}
	}

	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}

		glog.Error(errors.Wrapf(err, "while listing %s", pretty.Namespaces))
		return nil, gqlerror.New(err, pretty.Namespaces)
	}

	var ns []gqlschema.Namespace
	for _, n := range namespaces {
		ns = append(ns, gqlschema.Namespace{
			Name:   n.Name,
			Labels: n.Labels,
		})
	}

	return ns, nil
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

func (r *namespaceResolver) CreateNamespaceMutation(ctx context.Context, name string, labels gqlschema.Labels) (gqlschema.NamespaceCreationOutput, error) { //namespaceoutput

	// temporary: add 'env' label to newly created namespaces
	labels["env"] = "true"
	_, err := r.namespaceSvc.Create(name, labels)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s `%s`", pretty.Namespace, name))
		return gqlschema.NamespaceCreationOutput{}, gqlerror.New(err, pretty.Namespace, gqlerror.WithName(name))
	}
	return gqlschema.NamespaceCreationOutput{
		Name:   name,
		Labels: labels,
	}, nil
}

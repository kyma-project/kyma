package k8s

import (
	"context"

	"github.com/golang/glog"
	appPretty "github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/application/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/shared"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlerror"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/module"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: Write tests

//go:generate mockery -name=nsLister -output=automock -outpkg=automock -case=underscore
type nsLister interface {
	List() ([]v1.Namespace, error)
}

type namespaceResolver struct {
	nsLister     nsLister
	appRetriever shared.ApplicationRetriever
}

func newNamespaceResolver(nsLister nsLister, appRetriever shared.ApplicationRetriever) *namespaceResolver {
	return &namespaceResolver{
		nsLister:     nsLister,
		appRetriever: appRetriever,
	}
}

// TODO: Split this query into two
func (r *namespaceResolver) NamespacesQuery(ctx context.Context, applicationName *string) ([]gqlschema.Namespace, error) {
	var err error

	var namespaces []v1.Namespace
	if applicationName == nil {
		namespaces, err = r.nsLister.List()
	} else {
		var namespaceNames []string
		namespaceNames, err = r.appRetriever.Application().ListNamespacesFor(*applicationName)

		//TODO: Refactor 'ListNamespacesFor' to return []v1.Namespace and remove this workaround
		for _, nsName := range namespaceNames {
			namespaces = append(namespaces, v1.Namespace{
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
			Name: n.Name,
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

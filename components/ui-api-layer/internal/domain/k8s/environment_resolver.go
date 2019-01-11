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
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: Write tests

//go:generate mockery -name=envLister -output=automock -outpkg=automock -case=underscore
type envLister interface {
	List() ([]v1.Namespace, error)
}

type environmentResolver struct {
	envLister    envLister
	appRetriever shared.ApplicationRetriever
}

func newEnvironmentResolver(envLister envLister, appRetriever shared.ApplicationRetriever) *environmentResolver {
	return &environmentResolver{
		envLister:    envLister,
		appRetriever: appRetriever,
	}
}

// TODO: Split this query into two
func (r *environmentResolver) EnvironmentsQuery(ctx context.Context, applicationName *string) ([]gqlschema.Environment, error) {
	var err error

	var namespaces []v1.Namespace
	if applicationName == nil {
		namespaces, err = r.envLister.List()
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

		glog.Error(errors.Wrapf(err, "while listing %s", pretty.Environments))
		return nil, gqlerror.New(err, pretty.Environments)
	}

	var environments []gqlschema.Environment
	for _, ns := range namespaces {
		environments = append(environments, gqlschema.Environment{
			Name: ns.Name,
		})
	}

	return environments, nil
}

func (r *environmentResolver) ApplicationsField(ctx context.Context, obj *gqlschema.Environment) ([]string, error) {
	if obj == nil {
		return nil, errors.New("Cannot get application field for environment")
	}

	items, err := r.appRetriever.Application().ListInEnvironment(obj.Name)
	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}

		return nil, errors.Wrapf(err, "while listing %s for env", appPretty.Application)
	}

	var appNames []string
	for _, app := range items {
		appNames = append(appNames, app.Name)
	}

	return appNames, nil
}

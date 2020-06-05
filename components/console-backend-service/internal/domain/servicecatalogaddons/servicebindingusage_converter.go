package servicecatalogaddons

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/status"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/name"
	api "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	sbuTypes "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockery -name=statusBindingUsageExtractor -output=automock -outpkg=automock -case=underscore
type statusBindingUsageExtractor interface {
	Status(conditions []sbuTypes.ServiceBindingUsageCondition) gqlschema.ServiceBindingUsageStatus
}

//go:generate mockery -name=gqlServiceBindingUsageConverter -output=automock -outpkg=automock -case=underscore
type gqlServiceBindingUsageConverter interface {
	ToGQL(item *sbuTypes.ServiceBindingUsage) (*gqlschema.ServiceBindingUsage, error)
	ToGQLs(in []*sbuTypes.ServiceBindingUsage) ([]gqlschema.ServiceBindingUsage, error)
	InputToK8s(in *gqlschema.CreateServiceBindingUsageInput) (*api.ServiceBindingUsage, error)
}

type serviceBindingUsageConverter struct {
	extractor statusBindingUsageExtractor
}

func newBindingUsageConverter() serviceBindingUsageConverter {
	return serviceBindingUsageConverter{
		extractor: &status.BindingUsageExtractor{},
	}
}

func (c *serviceBindingUsageConverter) ToGQL(in *api.ServiceBindingUsage) (*gqlschema.ServiceBindingUsage, error) {
	if in == nil {
		return nil, nil
	}

	gqlSBU := gqlschema.ServiceBindingUsage{
		Name:      in.Name,
		Namespace: in.Namespace,
		UsedBy: gqlschema.LocalObjectReference{
			Name: in.Spec.UsedBy.Name,
			Kind: in.Spec.UsedBy.Kind,
		},
		ServiceBindingName: in.Spec.ServiceBindingRef.Name,
		Status:             c.extractor.Status(in.Status.Conditions),
	}

	if in.Spec.Parameters != nil && in.Spec.Parameters.EnvPrefix != nil {
		gqlSBU.Parameters = &gqlschema.ServiceBindingUsageParameters{
			EnvPrefix: &gqlschema.EnvPrefix{
				Name: in.Spec.Parameters.EnvPrefix.Name,
			},
		}
	}

	return &gqlSBU, nil
}

func (c *serviceBindingUsageConverter) ToGQLs(in []*api.ServiceBindingUsage) ([]gqlschema.ServiceBindingUsage, error) {
	var out []gqlschema.ServiceBindingUsage
	for _, u := range in {
		converted, err := c.ToGQL(u)
		if err != nil {
			return nil, err
		}

		if converted != nil {
			out = append(out, *converted)
		}
	}
	return out, nil
}

func (c *serviceBindingUsageConverter) InputToK8s(in *gqlschema.CreateServiceBindingUsageInput) (*api.ServiceBindingUsage, error) {
	if in == nil {
		return nil, nil
	}

	k8sSBU := api.ServiceBindingUsage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceBindingUsage",
			APIVersion: api.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name.EmptyIfNil(in.Name),
		},
		Spec: api.ServiceBindingUsageSpec{
			ServiceBindingRef: api.LocalReferenceByName{
				Name: in.ServiceBindingRef.Name,
			},
			UsedBy: api.LocalReferenceByKindAndName{
				Kind: in.UsedBy.Kind,
				Name: in.UsedBy.Name,
			},
		},
	}

	if in.Parameters != nil && in.Parameters.EnvPrefix != nil {
		k8sSBU.Spec.Parameters = &api.Parameters{
			EnvPrefix: &api.EnvPrefix{
				Name: in.Parameters.EnvPrefix.Name,
			},
		}
	}

	return &k8sSBU, nil
}

package servicecatalog

import (
	"fmt"

	api "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	sbuTypes "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/status"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
)

//go:generate mockery -name=statusBindingUsageExtractor -output=automock -outpkg=automock -case=underscore
type statusBindingUsageExtractor interface {
	Status(conditions []sbuTypes.ServiceBindingUsageCondition) gqlschema.ServiceBindingUsageStatus
}

type bindingUsageConverter struct {
	extractor statusBindingUsageExtractor
}

func newBindingUsageConverter() bindingUsageConverter {
	return bindingUsageConverter{
		extractor: &status.BindingUsageExtractor{},
	}
}

func (c *bindingUsageConverter) ToGQL(in *api.ServiceBindingUsage) (*gqlschema.ServiceBindingUsage, error) {
	if in == nil {
		return nil, nil
	}

	kind, err := c.refTypeToQL(in.Spec.UsedBy.Kind)
	if err != nil {
		return nil, err
	}

	gqlSBU := gqlschema.ServiceBindingUsage{
		Name:        in.Name,
		Environment: in.Namespace,
		UsedBy: gqlschema.LocalObjectReference{
			Name: in.Spec.UsedBy.Name,
			Kind: kind,
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

func (c *bindingUsageConverter) ToGQLs(in []*api.ServiceBindingUsage) ([]gqlschema.ServiceBindingUsage, error) {
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

func (c *bindingUsageConverter) InputToK8s(in *gqlschema.CreateServiceBindingUsageInput) (*api.ServiceBindingUsage, error) {
	if in == nil {
		return nil, nil
	}

	kind, err := c.referenceTypeToStr(in.UsedBy.Kind)
	if err != nil {
		return nil, err
	}

	k8sSBU := api.ServiceBindingUsage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceBindingUsage",
			APIVersion: api.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: in.Name,
		},
		Spec: api.ServiceBindingUsageSpec{
			ServiceBindingRef: api.LocalReferenceByName{
				Name: in.ServiceBindingRef.Name,
			},
			UsedBy: api.LocalReferenceByKindAndName{
				Kind: kind,
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

func (*bindingUsageConverter) referenceTypeToStr(referenceType gqlschema.BindingUsageReferenceType) (string, error) {
	switch referenceType {
	case gqlschema.BindingUsageReferenceTypeDeployment:
		return "Deployment", nil
	case gqlschema.BindingUsageReferenceTypeFunction:
		return "Function", nil
	default:
		return "", fmt.Errorf("unknown reference kind %s", referenceType)
	}
}

// refTypeToQL converts string to reference type, if the kind is unknown, returns exactly the same string.
func (*bindingUsageConverter) refTypeToQL(kind string) (gqlschema.BindingUsageReferenceType, error) {
	switch kind {
	case "Deployment":
		return gqlschema.BindingUsageReferenceTypeDeployment, nil
	case "Function":
		return gqlschema.BindingUsageReferenceTypeFunction, nil
	default:
		return "", fmt.Errorf("unknown kind %s", kind)
	}
}

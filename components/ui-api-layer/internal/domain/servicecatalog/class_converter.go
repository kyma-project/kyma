package servicecatalog

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/resource"
	"github.com/pkg/errors"
)

type classConverter struct{}

func (c *classConverter) ToGQL(in *v1beta1.ClusterServiceClass) (*gqlschema.ServiceClass, error) {
	if in == nil {
		return nil, nil
	}

	var externalMetadata map[string]interface{}
	var err error
	if in.Spec.ExternalMetadata != nil {
		externalMetadata, err = resource.ExtractRawToMap("ExternalMetadata", in.Spec.ExternalMetadata.Raw)
		if err != nil {
			return nil, errors.Wrapf(err, "while converting externalMetadata for ServiceClass `%s`", in.Name)
		}
	}

	providerDisplayName := resource.ToStringPtr(externalMetadata["providerDisplayName"])
	imageUrl := resource.ToStringPtr(externalMetadata["imageUrl"])
	documentationUrl := resource.ToStringPtr(externalMetadata["documentationUrl"])
	supportUrl := resource.ToStringPtr(externalMetadata["supportUrl"])
	displayName := resource.ToStringPtr(externalMetadata["displayName"])
	longDescription := resource.ToStringPtr(externalMetadata["longDescription"])

	class := gqlschema.ServiceClass{
		Name:                in.Name,
		ExternalName:        in.Spec.ExternalName,
		DisplayName:         displayName,
		Description:         in.Spec.Description,
		LongDescription:     longDescription,
		ProviderDisplayName: providerDisplayName,
		ImageUrl:            imageUrl,
		DocumentationUrl:    documentationUrl,
		SupportUrl:          supportUrl,
		CreationTimestamp:   in.CreationTimestamp.Time,
		Tags:                in.Spec.Tags,
	}

	return &class, nil
}

func (c *classConverter) ToGQLs(in []*v1beta1.ClusterServiceClass) ([]gqlschema.ServiceClass, error) {
	var result []gqlschema.ServiceClass
	for _, u := range in {
		converted, err := c.ToGQL(u)
		if err != nil {
			return nil, err
		}

		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result, nil
}

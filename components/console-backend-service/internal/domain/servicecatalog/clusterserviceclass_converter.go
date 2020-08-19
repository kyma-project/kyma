package servicecatalog

import (
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/pkg/errors"
)

type clusterServiceClassConverter struct{}

func (c *clusterServiceClassConverter) ToGQL(in *v1beta1.ClusterServiceClass) (*gqlschema.ClusterServiceClass, error) {
	if in == nil {
		return nil, nil
	}

	var externalMetadata map[string]interface{}
	var err error
	if in.Spec.ExternalMetadata != nil {
		externalMetadata, err = resource.ExtractRawToMap("ExternalMetadata", in.Spec.ExternalMetadata.Raw)
		if err != nil {
			return nil, errors.Wrapf(err, "while converting externalMetadata for ClusterServiceClass `%s`", in.Name)
		}
	}

	providerDisplayName := resource.ToStringPtr(externalMetadata["providerDisplayName"])
	imageURL := resource.ToStringPtr(externalMetadata["imageUrl"])
	documentationURL := resource.ToStringPtr(externalMetadata["documentationUrl"])
	supportURL := resource.ToStringPtr(externalMetadata["supportUrl"])
	displayName := resource.ToStringPtr(externalMetadata["displayName"])
	longDescription := resource.ToStringPtr(externalMetadata["longDescription"])

	labels := gqlschema.Labels{}
	err = labels.UnmarshalGQL(externalMetadata["labels"])
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling labels in ClusterServiceClass `%s`", in.Name)
	}

	class := gqlschema.ClusterServiceClass{
		Name:                in.Name,
		ExternalName:        in.Spec.ExternalName,
		DisplayName:         displayName,
		Description:         in.Spec.Description,
		LongDescription:     longDescription,
		ProviderDisplayName: providerDisplayName,
		ImageURL:            imageURL,
		DocumentationURL:    documentationURL,
		SupportURL:          supportURL,
		CreationTimestamp:   in.CreationTimestamp.Time,
		Tags:                in.Spec.Tags,
		Labels:              labels,
	}

	return &class, nil
}

func (c *clusterServiceClassConverter) ToGQLs(in []*v1beta1.ClusterServiceClass) ([]*gqlschema.ClusterServiceClass, error) {
	var result []*gqlschema.ClusterServiceClass
	for _, u := range in {
		converted, err := c.ToGQL(u)
		if err != nil {
			return nil, err
		}

		if converted != nil {
			result = append(result, converted)
		}
	}
	return result, nil
}

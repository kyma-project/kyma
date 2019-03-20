package k8s

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
)

type secretConverter struct{}

func (*secretConverter) ToGQL(in *v1.Secret) *gqlschema.Secret {
	if in == nil {
		return nil
	}

	out := &gqlschema.Secret{
		Name:         in.Name,
		Namespace:    in.ObjectMeta.Namespace,
		CreationTime: in.ObjectMeta.CreationTimestamp.Time,
		Type:         string(in.Type),
	}
	out.Data = make(gqlschema.JSON)
	for k, v := range in.Data {
		out.Data[k] = string(v)
	}

	out.Labels = make(gqlschema.JSON)
	for k, v := range in.ObjectMeta.Labels {
		out.Labels[k] = v
	}

	out.Annotations = make(gqlschema.JSON)

	for k, v := range in.ObjectMeta.Annotations {
		out.Annotations[k] = v
	}

	return out
}

func (s *secretConverter) ToGQLs(in []*v1.Secret) []gqlschema.Secret {
	var result []gqlschema.Secret
	for _, u := range in {
		converted := s.ToGQL(u)

		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result
}

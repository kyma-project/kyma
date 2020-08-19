package k8s

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/state"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

type replicaSetConverter struct {
	extractor state.ContainerExtractor
}

func (c *replicaSetConverter) ToGQL(in *apps.ReplicaSet) (*gqlschema.ReplicaSet, error) {
	if in == nil {
		return nil, nil
	}

	gqlJSON, err := c.replicaSetToGQLJSON(in)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting %s `%s` to it's json representation", pretty.ReplicaSet, in.Name)
	}
	labels := in.Labels
	if labels == nil {
		labels = gqlschema.Labels{}
	}

	return &gqlschema.ReplicaSet{
		Name:              in.Name,
		Pods:              c.getPods(in.Status),
		Namespace:         in.Namespace,
		Images:            c.getImages(in.Spec.Template.Spec.Containers),
		CreationTimestamp: in.CreationTimestamp.Time,
		Labels:            labels,
		JSON:              gqlJSON,
	}, nil
}

func (c *replicaSetConverter) ToGQLs(in []*apps.ReplicaSet) ([]*gqlschema.ReplicaSet, error) {
	var result []*gqlschema.ReplicaSet
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

func (c *replicaSetConverter) replicaSetToGQLJSON(in *apps.ReplicaSet) (gqlschema.JSON, error) {
	if in == nil {
		return nil, nil
	}

	jsonByte, err := json.Marshal(in)
	if err != nil {
		return nil, errors.Wrapf(err, "while marshalling %s `%s`", pretty.ReplicaSet, in.Name)
	}

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonByte, &jsonMap)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling %s `%s` to map", pretty.ReplicaSet, in.Name)
	}

	var result gqlschema.JSON
	err = result.UnmarshalGQL(jsonMap)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling %s `%s` to GQL JSON", pretty.ReplicaSet, in.Name)
	}

	return result, nil
}

func (c *replicaSetConverter) GQLJSONToReplicaSet(in gqlschema.JSON) (apps.ReplicaSet, error) {
	var buf bytes.Buffer
	in.MarshalGQL(&buf)
	bufBytes := buf.Bytes()
	result := apps.ReplicaSet{}
	err := json.Unmarshal(bufBytes, &result)
	if err != nil {
		return apps.ReplicaSet{}, errors.Wrapf(err, "while unmarshalling GQL JSON of %s", pretty.ReplicaSet)
	}

	return result, nil
}

func (c *replicaSetConverter) getPods(in apps.ReplicaSetStatus) string {
	return fmt.Sprintf("%d/%d", in.ReadyReplicas, in.Replicas)
}

func (c *replicaSetConverter) getImages(in []v1.Container) []string {
	var images []string
	for _, container := range in {
		images = append(images, container.Image)
	}
	return images
}

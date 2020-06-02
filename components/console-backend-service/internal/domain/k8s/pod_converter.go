package k8s

import (
	"bytes"
	"encoding/json"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/state"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
)

type podConverter struct {
	extractor state.ContainerExtractor
}

func (c *podConverter) ToGQL(in *v1.Pod) (*gqlschema.Pod, error) {
	if in == nil {
		return nil, nil
	}

	containerStates := c.extractor.States(in.Status.ContainerStatuses)

	gqlJSON, err := c.podToGQLJSON(in)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting %s `%s` to it's json representation", pretty.Pod, in.Name)
	}

	labels := in.Labels
	if labels == nil {
		labels = gqlschema.Labels{}
	}

	return &gqlschema.Pod{
		Name:              in.Name,
		NodeName:          in.Spec.NodeName,
		Namespace:         in.Namespace,
		RestartCount:      c.getRestartCount(in.Status.ContainerStatuses),
		CreationTimestamp: in.CreationTimestamp.Time,
		Labels:            labels,
		Status:            c.podStatusPhaseToGQLStatusType(in.Status.Phase),
		ContainerStates:   containerStates,
		JSON:              gqlJSON,
	}, nil
}

func (c *podConverter) ToGQLs(in []*v1.Pod) ([]*gqlschema.Pod, error) {
	var result []*gqlschema.Pod
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

func (c *podConverter) podToGQLJSON(in *v1.Pod) (gqlschema.JSON, error) {
	if in == nil {
		return nil, nil
	}

	jsonByte, err := json.Marshal(in)
	if err != nil {
		return nil, errors.Wrapf(err, "while marshalling %s `%s`", pretty.Pod, in.Name)
	}

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonByte, &jsonMap)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling %s `%s` to map", pretty.Pod, in.Name)
	}

	var result gqlschema.JSON
	err = result.UnmarshalGQL(jsonMap)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling %s `%s` to GQL JSON", pretty.Pod, in.Name)
	}

	return result, nil
}

func (c *podConverter) GQLJSONToPod(in gqlschema.JSON) (v1.Pod, error) {
	var buf bytes.Buffer
	in.MarshalGQL(&buf)
	bufBytes := buf.Bytes()
	result := v1.Pod{}
	err := json.Unmarshal(bufBytes, &result)
	if err != nil {
		return v1.Pod{}, errors.Wrapf(err, "while unmarshalling GQL JSON of %s", pretty.Pod)
	}

	return result, nil
}

func (c *podConverter) podStatusPhaseToGQLStatusType(in v1.PodPhase) gqlschema.PodStatusType {
	switch in {
	case v1.PodPending:
		return gqlschema.PodStatusTypePending
	case v1.PodRunning:
		return gqlschema.PodStatusTypeRunning
	case v1.PodSucceeded:
		return gqlschema.PodStatusTypeSucceeded
	case v1.PodFailed:
		return gqlschema.PodStatusTypeFailed
	default:
		return gqlschema.PodStatusTypeUnknown
	}
}

func (c *podConverter) getRestartCount(in []v1.ContainerStatus) int {
	restartCount := 0
	for _, containerStatus := range in {
		restartCount += int(containerStatus.RestartCount)
	}
	return restartCount
}

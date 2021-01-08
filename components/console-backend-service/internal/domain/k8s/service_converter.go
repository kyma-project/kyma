package k8s

import (
	"bytes"
	"encoding/json"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

type serviceConverter struct {
}

func (c *serviceConverter) ToGQL(in *v1.Service) (*gqlschema.Service, error) {
	if in == nil {
		return nil, nil
	}

	gqlJSON, err := c.serviceToGQLJSON(in)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting %s `%s` to it's json representation", pretty.Service, in.Name)
	}

	if in.Labels == nil {
		in.Labels = make(map[string]string)
	}

	return &gqlschema.Service{
		Name:              in.Name,
		ClusterIP:         in.Spec.ClusterIP,
		CreationTimestamp: in.CreationTimestamp.Time,
		Labels:            in.Labels,
		Ports:             toGQLSchemaServicePorts(in.Spec.Ports),
		Status:            toGQLSchemaServiceStatus(in.Status),
		JSON:              gqlJSON,
		UID:               string(in.ObjectMeta.UID),
	}, nil
}

func toGQLSchemaServiceStatus(s v1.ServiceStatus) *gqlschema.ServiceStatus {
	if s.LoadBalancer.Ingress == nil {
		return &gqlschema.ServiceStatus{
			LoadBalancer: &gqlschema.LoadBalancerStatus{
				Ingress: nil,
			},
		}
	}
	ingressSlice := make([]*gqlschema.LoadBalancerIngress, len(s.LoadBalancer.Ingress))
	for i, ingress := range s.LoadBalancer.Ingress {
		ingressSlice[i] = &gqlschema.LoadBalancerIngress{
			IP:       ingress.IP,
			HostName: ingress.Hostname,
		}
	}
	return &gqlschema.ServiceStatus{
		LoadBalancer: &gqlschema.LoadBalancerStatus{
			Ingress: ingressSlice,
		},
	}
}

func (c *serviceConverter) ToGQLs(in []*v1.Service) ([]*gqlschema.Service, error) {
	var result []*gqlschema.Service
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

func (c *serviceConverter) GQLJSONToService(in gqlschema.JSON) (v1.Service, error) {
	var buf bytes.Buffer
	in.MarshalGQL(&buf)
	bufBytes := buf.Bytes()
	result := v1.Service{}
	err := json.Unmarshal(bufBytes, &result)
	if err != nil {
		return v1.Service{}, errors.Wrapf(err, "while unmarshalling GQL JSON of %s", pretty.Service)
	}

	return result, nil
}

func (c *serviceConverter) serviceToGQLJSON(in *v1.Service) (gqlschema.JSON, error) {
	if in == nil {
		return nil, nil
	}

	jsonByte, err := json.Marshal(in)
	if err != nil {
		return nil, errors.Wrapf(err, "while marshalling %s `%s`", pretty.Service, in.Name)
	}

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonByte, &jsonMap)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling %s `%s` to map", pretty.Service, in.Name)
	}

	var result gqlschema.JSON
	err = result.UnmarshalGQL(jsonMap)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling %s `%s` to GQL JSON", pretty.Service, in.Name)
	}

	return result, nil
}

func toGQLSchemaServicePort(in *v1.ServicePort) *gqlschema.ServicePort {
	if in == nil {
		return nil
	}
	return &gqlschema.ServicePort{
		Name:            in.Name,
		ServiceProtocol: toGQLSchemaServiceProtocol(&in.Protocol),
		Port:            int(in.Port),
		NodePort:        int(in.NodePort),
		TargetPort:      int(in.TargetPort.IntVal),
	}
}

func toGQLSchemaServicePorts(in []v1.ServicePort) []*gqlschema.ServicePort {
	var result []*gqlschema.ServicePort
	for _, item := range in {
		converted := toGQLSchemaServicePort(&item)
		if converted != nil {
			result = append(result, converted)
		}
	}
	return result
}

func toGQLSchemaServiceProtocol(protocol *v1.Protocol) gqlschema.ServiceProtocol {
	switch *protocol {
	case v1.ProtocolTCP:
		return gqlschema.ServiceProtocolTCP
	case v1.ProtocolUDP:
		return gqlschema.ServiceProtocolUDP
	default:
		return gqlschema.ServiceProtocolUnknown
	}
}

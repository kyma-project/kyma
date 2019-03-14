package k8s

import (
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
)

type serviceConverter struct {
}

func (c *serviceConverter) ToGQL(in *v1.Service) *gqlschema.Service {
	if in == nil {
		return nil
	}
	return &gqlschema.Service{
		Name:              in.Name,
		ClusterIP:         in.Spec.ClusterIP,
		CreationTimestamp: in.CreationTimestamp.Time,
		Labels:            in.Labels,
		Ports:             toGQLSchemaServicePorts(in.Spec.Ports),
		Status:            toGQLSchemaServiceStatus(in.Status),
	}
}

func toGQLSchemaServiceStatus(s v1.ServiceStatus) gqlschema.ServiceStatus {
	var ingress []gqlschema.LoadBalancerIngress
	for _, i := range s.LoadBalancer.Ingress {
		ingress = append(ingress, gqlschema.LoadBalancerIngress{
			IP:       i.IP,
			HostName: i.Hostname,
		})
	}
	return gqlschema.ServiceStatus{
		LoadBalancer: gqlschema.LoadBalancerStatus{
			Ingress: ingress,
		},
	}
}

func (c *serviceConverter) ToGQLs(in []*v1.Service) []gqlschema.Service {
	var result []gqlschema.Service
	for _, u := range in {
		converted := c.ToGQL(u)
		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result
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

func toGQLSchemaServicePorts(in []v1.ServicePort) []gqlschema.ServicePort {
	var result []gqlschema.ServicePort
	for _, item := range in {
		converted := toGQLSchemaServicePort(&item)
		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result
}

func toGQLSchemaServiceProtocol(protocol *v1.Protocol) gqlschema.ServiceProtocol {
	switch *protocol {
	case v1.ProtocolTCP:
		return gqlschema.ServiceProtocolTcp
	case v1.ProtocolUDP:
		return gqlschema.ServiceProtocolUdp
	default:
		return gqlschema.ServiceProtocolUnknown
	}
}

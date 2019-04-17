package k8s

import (
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/authorization/v1"
)

type selfSubjectRulesConverter struct {
}

func (c *selfSubjectRulesConverter) ToGQL(in *v1.SelfSubjectRulesReview) (*gqlschema.SelfSubjectRules, error) {
	if in == nil {
		return nil, nil
	}

	fmt.Printf("NAmespace : %s", in.Spec.Namespace)
	fmt.Printf("Size : %d", len(in.Status.ResourceRules))
	fmt.Printf("Out : %+v", in.Status.ResourceRules)

	resourceRulesSlice := make([]*gqlschema.ResourceRule, len(in.Status.ResourceRules))
	for i, resourceRule := range in.Status.ResourceRules {
		resourceRulesSlice[i] = &gqlschema.ResourceRule{
			Verbs:     resourceRule.Verbs,
			APIGroups: resourceRule.APIGroups,
			Resources: resourceRule.Resources,
		}
	}

	out := &gqlschema.SelfSubjectRules{
		ResourceRules: resourceRulesSlice,
	}
	return out, nil
}

// func toGQLSchemaServiceStatus(s v1.ServiceStatus) gqlschema.ServiceStatus {
// 	if s.LoadBalancer.Ingress == nil {
// 		return gqlschema.ServiceStatus{
// 			LoadBalancer: gqlschema.LoadBalancerStatus{
// 				Ingress: nil,
// 			},
// 		}
// 	}
// 	ingressSlice := make([]gqlschema.LoadBalancerIngress, len(s.LoadBalancer.Ingress))
// 	for i, ingress := range s.LoadBalancer.Ingress {
// 		ingressSlice[i] = gqlschema.LoadBalancerIngress{
// 			IP:       ingress.IP,
// 			HostName: ingress.Hostname,
// 		}
// 	}
// 	return gqlschema.ServiceStatus{
// 		LoadBalancer: gqlschema.LoadBalancerStatus{
// 			Ingress: ingressSlice,
// 		},
// 	}
// }

// func (c *serviceConverter) ToGQLs(in []*v1.Service) ([]gqlschema.Service, error) {
// 	var result []gqlschema.Service
// 	for _, u := range in {
// 		converted, err := c.ToGQL(u)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if converted != nil {
// 			result = append(result, *converted)
// 		}
// 	}
// 	return result, nil
// }

// func (c *serviceConverter) GQLJSONToService(in gqlschema.JSON) (v1.Service, error) {
// 	var buf bytes.Buffer
// 	in.MarshalGQL(&buf)
// 	bufBytes := buf.Bytes()
// 	result := v1.Service{}
// 	err := json.Unmarshal(bufBytes, &result)
// 	if err != nil {
// 		return v1.Service{}, errors.Wrapf(err, "while unmarshalling GQL JSON of %s", pretty.Service)
// 	}

// 	return result, nil
// }

// func (c *serviceConverter) serviceToGQLJSON(in *v1.Service) (gqlschema.JSON, error) {
// 	if in == nil {
// 		return nil, nil
// 	}

// 	jsonByte, err := json.Marshal(in)
// 	if err != nil {
// 		return nil, errors.Wrapf(err, "while marshalling %s `%s`", pretty.Service, in.Name)
// 	}

// 	var jsonMap map[string]interface{}
// 	err = json.Unmarshal(jsonByte, &jsonMap)
// 	if err != nil {
// 		return nil, errors.Wrapf(err, "while unmarshalling %s `%s` to map", pretty.Service, in.Name)
// 	}

// 	var result gqlschema.JSON
// 	err = result.UnmarshalGQL(jsonMap)
// 	if err != nil {
// 		return nil, errors.Wrapf(err, "while unmarshalling %s `%s` to GQL JSON", pretty.Service, in.Name)
// 	}

// 	return result, nil
// }

// func toGQLSchemaServicePort(in *v1.ServicePort) *gqlschema.ServicePort {
// 	if in == nil {
// 		return nil
// 	}
// 	return &gqlschema.ServicePort{
// 		Name:            in.Name,
// 		ServiceProtocol: toGQLSchemaServiceProtocol(&in.Protocol),
// 		Port:            int(in.Port),
// 		NodePort:        int(in.NodePort),
// 		TargetPort:      int(in.TargetPort.IntVal),
// 	}
// }

// func toGQLSchemaServicePorts(in []v1.ServicePort) []gqlschema.ServicePort {
// 	var result []gqlschema.ServicePort
// 	for _, item := range in {
// 		converted := toGQLSchemaServicePort(&item)
// 		if converted != nil {
// 			result = append(result, *converted)
// 		}
// 	}
// 	return result
// }

// func toGQLSchemaServiceProtocol(protocol *v1.Protocol) gqlschema.ServiceProtocol {
// 	switch *protocol {
// 	case v1.ProtocolTCP:
// 		return gqlschema.ServiceProtocolTcp
// 	case v1.ProtocolUDP:
// 		return gqlschema.ServiceProtocolUdp
// 	default:
// 		return gqlschema.ServiceProtocolUnknown
// 	}
// }

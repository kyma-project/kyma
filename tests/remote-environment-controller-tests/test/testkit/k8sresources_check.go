package testkit
//
//import (
//	"testing"
//
//	istio "github.com/kyma-project/kyma/components/metadata-service/pkg/apis/istio/v1alpha2"
//	remoteenv "github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/remoteenvironment/v1alpha1"
//	"github.com/stretchr/testify/require"
//	v1core "k8s.io/api/core/v1"
//)
//
//func CheckK8sResources(t *testing.T, releaseName string) {
//	require.Equal(t, name, service.Name)
//
//	servicePorts := service.Spec.Ports[0]
//	require.Equal(t, protocol, servicePorts.Protocol)
//	require.Equal(t, int32(port), servicePorts.Port)
//	require.Equal(t, int32(targetPort), servicePorts.TargetPort.IntVal)
//
//	checkLabels(t, labels, service.Labels)
//}
//
//
//func checkLabels(t *testing.T, expected, actual Labels) {
//	for key := range expected {
//		require.Equal(t, expected[key], actual[key])
//	}
//}
//
//func makeMatchExpression(name, namespace string) string {
//	return `(destination.service == "` + name + "." + namespace + `.svc.cluster.local") && (source.labels["` + name + `"] != "true")`
//}
//
//func findServiceInRemoteEnv(reServices []remoteenv.Service, searchedID string) *remoteenv.Service {
//	for _, e := range reServices {
//		if e.ID == searchedID {
//			return &e
//		}
//	}
//	return nil
//}
//
//func findEntryOfType(entries []remoteenv.Entry, typeName string) *remoteenv.Entry {
//	for _, e := range entries {
//		if e.Type == typeName {
//			return &e
//		}
//	}
//	return nil
//}

package resources

import (
	"fmt"
	"strings"

	"k8s.io/client-go/util/cert"
)

func GenerateWebhookCertificates(serviceName, namespace string) ([]byte, []byte, error) {
	namespacedServiceName := strings.Join([]string{serviceName, namespace}, ".")
	commonName := strings.Join([]string{namespacedServiceName, "svc"}, ".")
	serviceHostname := fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, namespace)

	serviceNames := []string{
		serviceName,
		namespacedServiceName,
		commonName,
		serviceHostname,
	}
	return cert.GenerateSelfSignedCertKey(commonName, nil, serviceNames)
}

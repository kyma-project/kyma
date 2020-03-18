package main

import (
	"fmt"
	"strings"
	"time"

	"kyma-project.io/compass-runtime-agent/internal/compass/director"

	"k8s.io/apimachinery/pkg/types"
)

const (
	defaultNamespace = "default"
)

type Config struct {
	ConnectionConfigMap            string        `envconfig:"default=compass-system/compass-agent-configuration"`
	ControllerSyncPeriod           time.Duration `envconfig:"default=20s"`
	MinimalCompassSyncTime         time.Duration `envconfig:"default=10s"`
	CertValidityRenewalThreshold   float64       `envconfig:"default=0.3"`
	ClusterCertificatesSecret      string        `envconfig:"default=compass-system/cluster-client-certificates"`
	CaCertificatesSecret           string        `envconfig:"default=istio-system/ca-certificates"`
	InsecureConnectorCommunication bool          `envconfig:"default=false"`
	IntegrationNamespace           string        `envconfig:"default=kyma-integration"`
	GatewayPort                    int           `envconfig:"default=8080"`
	InsecureConfigurationFetch     bool          `envconfig:"default=false"`
	UploadServiceUrl               string        `envconfig:"default=http://rafter-upload-service.kyma-system.svc.cluster.local:80"`
	QueryLogging                   bool          `envconfig:"default=false"`
	DirectorProxy                  director.ProxyConfig
	MetricsLoggingTimeInterval     time.Duration `envconfig:"default=30m"`
	//EnableApiPackages              bool          `envconfig:"default=false"`

	Runtime director.RuntimeURLsConfig
}

func (o *Config) String() string {
	return fmt.Sprintf("ConnectionConfigMap=%s, "+
		"ControllerSyncPeriod=%s, MinimalCompassSyncTime=%s, "+
		"CertValidityRenewalThreshold=%f, ClusterCertificatesSecret=%s, CaCertificatesSecret=%s, "+
		"IntegrationNamespace=%s, GatewayPort=%d, InsecureConfigurationFetch=%v, UploadServiceUrl=%s, "+
		"QueryLogging=%v, MetricsLoggingTimeInterval=%s, "+
		"RuntimeEventsURL=%s, RuntimeConsoleURL=%s"+
		"DirectorProxyPort=%v,  DirectorProxyInsecureSkipVerify=%v",
		o.ConnectionConfigMap,
		o.ControllerSyncPeriod.String(), o.MinimalCompassSyncTime.String(),
		o.CertValidityRenewalThreshold, o.ClusterCertificatesSecret, o.CaCertificatesSecret,
		o.IntegrationNamespace, o.GatewayPort, o.InsecureConfigurationFetch, o.UploadServiceUrl,
		o.QueryLogging, o.MetricsLoggingTimeInterval,
		o.Runtime.EventsURL, o.Runtime.ConsoleURL,
		o.DirectorProxy.Port, o.DirectorProxy.InsecureSkipVerify)
}

func parseNamespacedName(value string) types.NamespacedName {
	parts := strings.Split(value, string(types.Separator))

	if singleValueProvided(parts) {
		return types.NamespacedName{
			Name:      parts[0],
			Namespace: defaultNamespace,
		}
	}

	namespace := get(parts, 0)
	if namespace == "" {
		namespace = defaultNamespace
	}

	return types.NamespacedName{
		Namespace: namespace,
		Name:      get(parts, 1),
	}
}

func singleValueProvided(split []string) bool {
	return len(split) == 1 || get(split, 1) == ""
}

func get(array []string, index int) string {
	if len(array) > index {
		return array[index]
	}
	return ""
}

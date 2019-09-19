package main

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/types"
)

const (
	defaultNamespace = "default"
)

// TODO - This will be removed in favour of mounting Config Map
type EnvConfig struct {
	//DirectorURL string `envconfig:"DIRECTOR_URL"`
	ConnectorURL string `envconfig:"CONNECTOR_URL"`
	Token        string `envconfig:"TOKEN"`
	//RuntimeId    string `envconfig:"RUNTIME_ID"`
	//Tenant       string `envconfig:"TENANT"`
}

type Config struct {
	ConfigFile                     string        `envconfig:"default=/etc/config/config.json"`
	ControllerSyncPeriod           time.Duration `envconfig:"default=60s"`
	MinimalCompassSyncTime         time.Duration `envconfig:"default=300s"`
	CertValidityRenewalThreshold   float64       `envconfig:"default=0.3"`
	ClusterCertificatesSecret      string        `envconfig:"default=kyma-integration/cluster-client-certificates"`
	CaCertificatesSecret           string        `envconfig:"default=istio-system/ca-certificates"`
	InsecureConnectorCommunication bool          `envconfig:"default=false"`
	IntegrationNamespace           string        `envconfig:"default=kyma-integration"`
	GatewayPort                    int           `envconfig:"default=8080"`
	InsecureConfigurationFetch     bool          `envconfig:"default=false"`
	UploadServiceUrl               string        `envconfig:"default=http://assetstore-asset-upload-service.kyma-system.svc.cluster.local:3000"`
}

func (o *Config) String() string {
	return fmt.Sprintf("ConfigFile=%s, "+
		"ControllerSyncPeriod=%d, MinimalCompassSyncTime=%d, "+
		"CertValidityRenewalThreshold=%f, ClusterCertificatesSecret=%s, CaCertificatesSecret=%s, "+
		"IntegrationNamespace=%s, GatewayPort=%d, InsecureConfigurationFetch=%v, UploadServiceUrl=%s",
		o.ConfigFile,
		o.ControllerSyncPeriod, o.MinimalCompassSyncTime,
		o.CertValidityRenewalThreshold, o.ClusterCertificatesSecret, o.CaCertificatesSecret,
		o.IntegrationNamespace, o.GatewayPort, o.InsecureConfigurationFetch, o.UploadServiceUrl)
}

// TODO - This will be removed in favour of mounting Config Map
func (ec EnvConfig) String() string {
	return fmt.Sprintf("CONNECTOR_URL=%s, TOKEN_PROVIDED=%t", ec.ConnectorURL, ec.Token != "")
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

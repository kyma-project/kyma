package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass/director"

	"k8s.io/apimachinery/pkg/types"
)

const (
	defaultNamespace = "default"
)

type Config struct {
	AgentConfigurationSecret           string        `envconfig:"default=kyma-system/compass-agent-configuration"`
	ControllerSyncPeriod               time.Duration `envconfig:"default=20s"`
	MinimalCompassSyncTime             time.Duration `envconfig:"default=10s"`
	CertValidityRenewalThreshold       float64       `envconfig:"default=0.3"`
	ClusterCertificatesSecret          string        `envconfig:"default=kyma-system/cluster-client-certificates"`
	CaCertificatesSecret               string        `envconfig:"default=istio-system/ca-certificates"`
	SkipCompassTLSVerify               bool          `envconfig:"default=false"`
	GatewayPort                        int           `envconfig:"default=8080"`
	SkipAppsTLSVerify                  bool          `envconfig:"default=false"`
	CentralGatewayServiceUrl           string        `envconfig:"default=http://central-application-gateway.kyma-system.svc.cluster.local:8082"`
	QueryLogging                       bool          `envconfig:"default=false"`
	MetricsLoggingTimeInterval         time.Duration `envconfig:"default=30m"`
	HealthPort                         string        `envconfig:"default=8090"`
	IntegrationNamespace               string        `envconfig:"default=kyma-system"`
	CaCertSecretToMigrate              string        `envconfig:"optional"`
	CaCertSecretKeysToMigrate          string        `envconfig:"default='cacert'"`
	ClusterCertificatesSecretToMigrate string        `envconfig:"optional"`
	AgentConfigurationSecretToMigrate  string        `envconfig:"optional"`
	Runtime                            director.RuntimeURLsConfig
}

func (o *Config) String() string {
	return fmt.Sprintf("AgentConfigurationSecret=%s, "+
		"ControllerSyncPeriod=%s, MinimalCompassSyncTime=%s, "+
		"CertValidityRenewalThreshold=%f, ClusterCertificatesSecret=%s, CaCertificatesSecret=%s, "+
		"SkipCompassTLSVerify=%v, GatewayPort=%d,"+
		"SkipAppTLSVerify=%v, "+
		"QueryLogging=%v, MetricsLoggingTimeInterval=%s, "+
		"RuntimeEventsURL=%s, RuntimeConsoleURL=%s, "+
		"HealthPort=%s, IntegrationNamespace=%s, CaCertSecretToMigrate=%s, caCertificateSecretKeysToMigrate=%s, "+
		"ClusterCertificatesSecretToMigrate=%s, AgentConfigurationSecretToMigrate=%s, "+
		"CentralGatewayServiceUrl=%v",
		o.AgentConfigurationSecret,
		o.ControllerSyncPeriod.String(), o.MinimalCompassSyncTime.String(),
		o.CertValidityRenewalThreshold, o.ClusterCertificatesSecret, o.CaCertificatesSecret,
		o.SkipCompassTLSVerify, o.GatewayPort,
		o.SkipAppsTLSVerify,
		o.QueryLogging, o.MetricsLoggingTimeInterval,
		o.Runtime.EventsURL, o.Runtime.ConsoleURL,
		o.HealthPort, o.IntegrationNamespace, o.CaCertSecretToMigrate, o.CaCertSecretKeysToMigrate,
		o.ClusterCertificatesSecretToMigrate, o.AgentConfigurationSecretToMigrate,
		o.CentralGatewayServiceUrl)
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

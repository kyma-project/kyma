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
	AgentConfigurationSecret     string        `envconfig:"default=compass-system/compass-agent-configuration"`
	ControllerSyncPeriod         time.Duration `envconfig:"default=20s"`
	MinimalCompassSyncTime       time.Duration `envconfig:"default=10s"`
	TimeToRequeueAfterFailure    time.Duration `envconfig:"default=2m"`
	CertValidityRenewalThreshold float64       `envconfig:"default=0.3"`
	ClusterCertificatesSecret    string        `envconfig:"default=compass-system/cluster-client-certificates"`
	CaCertificatesSecret         string        `envconfig:"default=istio-system/ca-certificates"`
	SkipCompassTLSVerify         bool          `envconfig:"default=false"`
	GatewayPort                  int           `envconfig:"default=8080"`
	UploadServiceUrl             string        `envconfig:"default=http://rafter-upload-service.kyma-system.svc.cluster.local:80"`
	QueryLogging                 bool          `envconfig:"default=false"`
	DirectorProxy                director.ProxyConfig
	MetricsLoggingTimeInterval   time.Duration `envconfig:"default=30m"`
	HealthPort                   string        `envconfig:"default=8090"`

	Runtime director.RuntimeURLsConfig
}

func (o *Config) String() string {
	return fmt.Sprintf("AgentConfigurationSecret=%s, "+
		"ControllerSyncPeriod=%s, MinimalCompassSyncTime=%s, TimeToRequeueAfterFailure=%s"+
		"CertValidityRenewalThreshold=%f, ClusterCertificatesSecret=%s, CaCertificatesSecret=%s, "+
		"SkipCompassTLSVerify=%v, GatewayPort=%d, UploadServiceUrl=%s, "+
		"QueryLogging=%v, MetricsLoggingTimeInterval=%s, "+
		"RuntimeEventsURL=%s, RuntimeConsoleURL=%s"+
		"DirectorProxyPort=%v,  DirectorProxyInsecureSkipVerify=%v, HealthPort=%s",
		o.AgentConfigurationSecret,
		o.ControllerSyncPeriod.String(), o.MinimalCompassSyncTime.String(), o.TimeToRequeueAfterFailure.String(),
		o.CertValidityRenewalThreshold, o.ClusterCertificatesSecret, o.CaCertificatesSecret,
		o.SkipCompassTLSVerify, o.GatewayPort, o.UploadServiceUrl,
		o.QueryLogging, o.MetricsLoggingTimeInterval,
		o.Runtime.EventsURL, o.Runtime.ConsoleURL,
		o.DirectorProxy.Port, o.DirectorProxy.InsecureSkipVerify, o.HealthPort)
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

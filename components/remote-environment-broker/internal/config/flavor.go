package config

import (
	"fmt"
	"regexp"
)

// BrokerFlavor aggregates information about mode REB is running: if it is cluster scoped or namespaced scoped
type BrokerFlavor struct {
	clusterScoped      bool
	clusterBrokerName  string
	nsBrokerURLPattern *regexp.Regexp
}

// NewBrokerFlavorFromConfig creates BrokerFlavor from configuration
func NewBrokerFlavorFromConfig(cfg *Config) *BrokerFlavor {
	return &BrokerFlavor{
		clusterBrokerName:  cfg.ClusterScopedBrokerName,
		clusterScoped:      cfg.ClusterScopedBroker,
		nsBrokerURLPattern: regexp.MustCompile("reb-ns-for-([a-z][a-z0-9-]*)\\."),
	}
}

// GetNsFromBrokerURL extracts namespace from broker URL
func (bf *BrokerFlavor) GetNsFromBrokerURL(url string) (string, error) {
	out := bf.nsBrokerURLPattern.FindStringSubmatch(url)
	if len(out) == 0 {
		return "", fmt.Errorf("url:%s does not match pattern %s", url, bf.nsBrokerURLPattern.String())
	}
	// out[0] = matched regexp, out[1] = matched group in bracket
	return out[1], nil
}

// IsClusterScoped returns information if broker is configured as a ClusterScoped
func (bf *BrokerFlavor) IsClusterScoped() bool {
	return bf.clusterScoped
}

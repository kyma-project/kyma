package mode

import (
	"fmt"
	"regexp"

	"github.com/kyma-project/kyma/components/remote-environment-broker/internal/config"
	"github.com/pkg/errors"
)

// BrokerService aggregates information about mode REB is running: if it is cluster scoped or namespaced scoped
type BrokerService struct {
	clusterScoped      bool
	clusterBrokerName  string
	nsBrokerURLPattern *regexp.Regexp
}

// NewBrokerService creates BrokerService from configuration
func NewBrokerService(cfg *config.Config) (*BrokerService, error) {
	r, err := regexp.Compile("reb-ns-for-([a-z][a-z0-9-]*)\\.")
	if err != nil {
		return nil, errors.Wrap(err, "while compiling regexp for URL of namespaced brokers")
	}
	return &BrokerService{
		clusterBrokerName:  cfg.ClusterScopedBrokerName,
		clusterScoped:      cfg.ClusterScopedBrokerEnabled,
		nsBrokerURLPattern: r,
	}, nil

}

// GetNsFromBrokerURL extracts namespace from broker URL
func (bf *BrokerService) GetNsFromBrokerURL(url string) (string, error) {
	out := bf.nsBrokerURLPattern.FindStringSubmatch(url)
	if len(out) < 2 {
		return "", fmt.Errorf("url:%s does not match pattern %s", url, bf.nsBrokerURLPattern.String())
	}
	// out[0] = matched regexp, out[1] = matched group in bracket
	return out[1], nil
}

// IsClusterScoped returns information if broker is configured as a ClusterScoped
func (bf *BrokerService) IsClusterScoped() bool {
	return bf.clusterScoped
}

package broker

import (
	"fmt"
	"regexp"

	"github.com/pkg/errors"
)

const prefix = "reb-ns-for-"

// NsBrokerService provides information about REB brokers
type NsBrokerService struct {
	nsBrokerURLPattern *regexp.Regexp
}

// NewNsBrokerService creates NsBrokerService from configuration
func NewNsBrokerService() (*NsBrokerService, error) {
	r, err := regexp.Compile(fmt.Sprintf("%s([a-z][a-z0-9-]*)\\.", prefix))
	if err != nil {
		return nil, errors.Wrap(err, "while compiling regexp for URL of namespaced brokers")
	}
	return &NsBrokerService{
		nsBrokerURLPattern: r,
	}, nil

}

// GetNsFromBrokerURL extracts namespace from broker URL
func (bf *NsBrokerService) GetNsFromBrokerURL(url string) (string, error) {
	out := bf.nsBrokerURLPattern.FindStringSubmatch(url)
	if len(out) < 2 {
		return "", fmt.Errorf("url:%s does not match pattern %s", url, bf.nsBrokerURLPattern.String())
	}
	// out[0] = matched regexp, out[1] = matched group in bracket
	return out[1], nil
}

// GetServiceNameForNsBroker returns service name for namespaced broker
func (bf *NsBrokerService) GetServiceNameForNsBroker(ns string) string {
	return prefix + ns
}

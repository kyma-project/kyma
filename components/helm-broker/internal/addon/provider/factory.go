package provider

import (
	"fmt"
	"io"
	"net/url"
	"regexp"

	getter "github.com/hashicorp/go-getter"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

// Provider defines factory func for returning concrete addon RepositoryGetter
type Provider func(idxAddr, dstPath string) (RepositoryGetter, error)

// addonLoader provides function for loading addon both from archive and directory.
type addonLoader interface {
	Load(io.Reader) (*internal.Addon, []*chart.Chart, error)
	LoadDir(path string) (*internal.Addon, []*chart.Chart, error)
}

// ClientFactory knows how to build the concrete RepositoryGetter for given addon repository URL.
type ClientFactory struct {
	gettersProviders    map[string]Provider
	specifiedSchemRegex *regexp.Regexp
	log                 *logrus.Entry
	addonLoader         addonLoader
}

// NewClientFactory returns new instance of the ClientFactory
func NewClientFactory(allowedGetters map[string]Provider, addonLoader addonLoader, log logrus.FieldLogger) (*ClientFactory, error) {
	specifiedSchemRegex, err := regexp.Compile(`^([A-Za-z0-9]+)::(.+)$`)
	if err != nil {
		return nil, err
	}
	return &ClientFactory{
		specifiedSchemRegex: specifiedSchemRegex,
		gettersProviders:    allowedGetters,
		addonLoader:         addonLoader,
		log:                 log.WithField("service", "addonClientFactory"),
	}, nil
}

// NewGetter decodes and returns new concrete RepositoryGetter for given type of the repository URL.
func (cli *ClientFactory) NewGetter(rawURL, instPath string) (AddonClient, error) {
	// normalization
	fullRealAddr, err := getter.Detect(rawURL, instPath, getter.Detectors)
	if err != nil {
		return nil, err
	}

	if fullRealAddr != rawURL {
		cli.log.Infof("[TRACE] go-getter detectors rewrote %q to %q", rawURL, fullRealAddr)
	}

	// get schema + source address
	fullRealAddrURL, err := url.Parse(fullRealAddr)
	if err != nil {
		return nil, err
	}

	// get concrete RepositoryGetter provider
	scheme, realAddr := cli.getSchemaAndSrc(fullRealAddrURL)
	getterProvider, ok := cli.gettersProviders[scheme]
	if !ok {
		return nil, fmt.Errorf("not supported scheme '%s' for addons repository", scheme)
	}

	// create concrete RepositoryGetter
	concreteGetter, err := getterProvider(realAddr, instPath)
	if err != nil {
		return nil, err
	}

	addonClient, err := NewClient(concreteGetter, cli.addonLoader, cli.log.WithField("getterScheme", scheme))
	if err != nil {
		return nil, err
	}

	return addonClient, nil
}

// forcedRegexp is the regular expression that finds forced gettersProviders. This
// syntax is schema::url, example: git::https://foo.com
//
// Copied from: https://github.com/hashicorp/go-getter/blob/0be63f2a663a793576db30b67ba5fa79f54c1afb/get.go#L57
// because it was unexported. We can consider contribution in future to support our case.
var forcedRegexp = regexp.MustCompile(`^([A-Za-z0-9]+)::(.+)$`)

// getSchemaAndSrc takes a source and returns the tuple of the RepositoryGetter schema
// and the raw URL (without the force syntax).
func (cli *ClientFactory) getSchemaAndSrc(url *url.URL) (string, string) {
	if ms := forcedRegexp.FindStringSubmatch(url.String()); ms != nil {
		schem, src := ms[1], ms[2]
		return schem, src
	}
	return url.Scheme, url.String()
}

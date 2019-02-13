package bundle

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

//go:generate mockery -name=bundleRemover -output=automock -outpkg=automock -case=underscore

type bundleRemover interface {
	RemoveAll() error
}

// Updater is responsible for updating repositories URL on configmap change
type Updater struct {
	cacheURL      string
	bundleRemover bundleRemover
	log           *logrus.Entry
}

// NewUpdater returns new Updater instance.
func NewUpdater(bundleRemover bundleRemover, log logrus.FieldLogger) *Updater {
	return &Updater{
		bundleRemover: bundleRemover,
		log:           log.WithField("service", "updater"),
	}
}

// SwitchRepositories is executed when the repositories url changed.
func (u *Updater) SwitchRepositories(url string) ([]RepositoryConfig, error) {
	err := u.cleanRepositories()
	if err != nil {
		return nil, errors.Wrap(err, "while cleaning bundles repositories")
	}
	u.saveURL(url)

	return u.repositoryConfigs(), nil
}

// IsURLChanged returns true if URL changed since last update.
func (u *Updater) IsURLChanged(url string) bool {
	return u.cacheURL != url
}

func (u *Updater) repositoryConfigs() []RepositoryConfig {
	var cfgs []RepositoryConfig
	for _, url := range strings.Split(u.cacheURL, ";") {
		cfgs = append(cfgs, RepositoryConfig{
			URL: url,
		})
	}
	return cfgs
}

func (u *Updater) cleanRepositories() error {
	if err := u.bundleRemover.RemoveAll(); err != nil {
		return errors.Wrap(err, "while removing all bundles on url change")
	}
	return nil
}

func (u *Updater) saveURL(url string) {
	u.cacheURL = url
}

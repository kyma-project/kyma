package controller

import (
	"github.com/Masterminds/semver"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/addon"
	"github.com/kyma-project/kyma/components/helm-broker/internal/addon/provider"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

//go:generate mockery -name=addonStorage -output=automock -outpkg=automock -case=underscore
type addonStorage interface {
	Get(internal.Namespace, internal.AddonName, semver.Version) (*internal.Addon, error)
	Upsert(internal.Namespace, *internal.Addon) (replace bool, err error)
	Remove(internal.Namespace, internal.AddonName, semver.Version) error
	FindAll(internal.Namespace) ([]*internal.Addon, error)
}

//go:generate mockery -name=chartStorage -output=automock -outpkg=automock -case=underscore
type chartStorage interface {
	Upsert(internal.Namespace, *chart.Chart) (replace bool, err error)
	Remove(internal.Namespace, internal.ChartName, semver.Version) error
}

//go:generate mockery -name=addonGetterFactory -output=automock -outpkg=automock -case=underscore
type addonGetterFactory interface {
	NewGetter(rawURL, instPath string) (provider.AddonClient, error)
}

//go:generate mockery -name=addonGetter -output=automock -outpkg=automock -case=underscore
type addonGetter interface {
	Cleanup() error
	GetCompleteAddon(entry addon.EntryDTO) (addon.CompleteAddon, error)
	GetIndex() (*addon.IndexDTO, error)
}

//go:generate mockery -name=brokerFacade -output=automock -outpkg=automock -case=underscore
type brokerFacade interface {
	Create(ns string) error
	Exist(ns string) (bool, error)
	Delete(ns string) error
}

//go:generate mockery -name=docsProvider -output=automock -outpkg=automock -case=underscore
type docsProvider interface {
	EnsureDocsTopic(addon *internal.Addon, namespace string) error
	EnsureDocsTopicRemoved(id string, namespace string) error
}

//go:generate mockery -name=brokerSyncer -output=automock -outpkg=automock -case=underscore
type brokerSyncer interface {
	SyncServiceBroker(namespace string) error
}

//go:generate mockery -name=clusterBrokerFacade -output=automock -outpkg=automock -case=underscore
type clusterBrokerFacade interface {
	Create() error
	Exist() (bool, error)
	Delete() error
}

//go:generate mockery -name=clusterDocsProvider -output=automock -outpkg=automock -case=underscore
type clusterDocsProvider interface {
	EnsureClusterDocsTopic(addon *internal.Addon) error
	EnsureClusterDocsTopicRemoved(id string) error
}

//go:generate mockery -name=clusterBrokerSyncer -output=automock -outpkg=automock -case=underscore
type clusterBrokerSyncer interface {
	Sync() error
}

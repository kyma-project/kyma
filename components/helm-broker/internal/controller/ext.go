package controller

import (
	"github.com/Masterminds/semver"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

//go:generate mockery -name=bundleStorage -output=automock -outpkg=automock -case=underscore
type bundleStorage interface {
	Get(internal.Namespace, internal.BundleName, semver.Version) (*internal.Bundle, error)
	Upsert(internal.Namespace, *internal.Bundle) (replace bool, err error)
	Remove(internal.Namespace, internal.BundleName, semver.Version) error
	FindAll(internal.Namespace) ([]*internal.Bundle, error)
}

//go:generate mockery -name=chartStorage -output=automock -outpkg=automock -case=underscore
type chartStorage interface {
	Upsert(internal.Namespace, *chart.Chart) (replace bool, err error)
	Remove(internal.Namespace, internal.ChartName, semver.Version) error
}

//go:generate mockery -name=bundleProvider -output=automock -outpkg=automock -case=underscore
type bundleProvider interface {
	GetIndex(string) (*bundle.IndexDTO, error)
	LoadCompleteBundle(bundle.EntryDTO) (bundle.CompleteBundle, error)
}

//go:generate mockery -name=brokerFacade -output=automock -outpkg=automock -case=underscore
type brokerFacade interface {
	Create(ns string) error
	Exist(ns string) (bool, error)
	Delete(ns string) error
}

//go:generate mockery -name=docsProvider -output=automock -outpkg=automock -case=underscore
type docsProvider interface {
	EnsureDocsTopic(bundle *internal.Bundle, namespace string) error
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
	EnsureClusterDocsTopic(bundle *internal.Bundle) error
	EnsureClusterDocsTopicRemoved(id string) error
}

//go:generate mockery -name=clusterBrokerSyncer -output=automock -outpkg=automock -case=underscore
type clusterBrokerSyncer interface {
	Sync() error
}

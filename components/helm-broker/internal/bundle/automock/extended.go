package automock

import (
	"io"
	"strings"

	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
)

func (m *Repository) ExpectOnIndexReader(idxContent string) {
	r := strings.NewReader(idxContent)
	m.On("IndexReader").Return(r, func() {}, nil)
}

func (m *Repository) ExpectErrorOnIndexReader(outErr error) {
	m.On("IndexReader").Return(nil, nil, outErr)
}

func (m *Repository) ExpectOnBundleReader(givenName bundle.Name, givenVersion bundle.Version, bundleReader io.Reader) {
	m.On("BundleReader", givenName, givenVersion).Return(bundleReader, func() {}, nil).Once()
}

func (m *Repository) ExpectErrorOnBundleReader(givenName bundle.Name, givenVersion bundle.Version, outErr error) {
	m.On("BundleReader", givenName, givenVersion).Return(nil, nil, outErr).Once()
}

func (mati *BundleLoader) ExpectOnLoad(r io.Reader, outBundle *internal.Bundle, outCharts []*chart.Chart) {
	mati.On("Load", r).Return(outBundle, outCharts, nil).Once()
}

func (mati *BundleLoader) ExpectErrorOnLoad(r io.Reader, outErr error) {
	mati.On("Load", r).Return(nil, nil, outErr).Once()
}

func (bi *BundleUpserter) ExpectOnUpsert(inBundle *internal.Bundle) {
	bi.On("Upsert", inBundle).Return(false, nil).Once()
}

func (bi *BundleUpserter) ExpectErrorOnUpsert(inBundle *internal.Bundle, outErr error) {
	bi.On("Upsert", inBundle).Return(false, outErr).Once()
}

func (ci *ChartUpserter) ExpectOnUpsert(inChart *chart.Chart) {
	ci.On("Upsert", inChart).Return(false, nil).Once()
}

func (ci *ChartUpserter) ExpectErrorOnUpsert(inChart *chart.Chart, outErr error) {
	ci.On("Upsert", inChart).Return(false, outErr).Once()
}

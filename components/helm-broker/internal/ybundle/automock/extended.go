package automock

import (
	"strings"

	"github.com/stretchr/testify/mock"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"io"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/ybundle"
)

func (m *Repository) ExpectOnIndexReader(idxContent string) {
	r := strings.NewReader(idxContent)
	m.On("IndexReader").Return(r, func() {}, nil)
}

func (m *Repository) ExpectErrorOnIndexReader(outErr error) {
	m.On("IndexReader").Return(nil, nil, outErr)
}

func (m *Repository) ExpectOnBundleReader(givenName ybundle.BundleName, givenVersion ybundle.BundleVersion, bundleReader io.Reader) {
	m.On("BundleReader", givenName, givenVersion).Return(bundleReader, func() {}, nil)
}

func (m *Repository) ExpectErrorOnBundleReader(outErr error) {
	m.On("BundleReader", mock.Anything, mock.Anything).Return(nil, nil, outErr)
}

func (mati *BundleLoader) ExpectOnLoad(r io.Reader, outBundle *internal.Bundle, outCharts []*chart.Chart) {
	mati.On("Load", r).Return(outBundle, outCharts, nil)
}

func (mati *BundleLoader) ExpectErrorOnLoad(outErr error) {
	mati.On("Load", mock.Anything).Return(nil, nil, outErr)
}

func (bi *BundleUpserter) ExpectOnUpsert(inBundle *internal.Bundle) {
	bi.On("Upsert", inBundle).Return(false, nil)
}

func (bi *BundleUpserter) ExpectErrorOnUpsert(outErr error) {
	bi.On("Upsert", mock.Anything).Return(false, outErr)
}

func (ci *ChartUpserter) ExpectOnUpsert(inChart *chart.Chart) {
	ci.On("Upsert", inChart).Return(false, nil)
}

func (ci *ChartUpserter) ExpectErrorOnUpsert(outErr error) {
	ci.On("Upsert", mock.Anything).Return(false, outErr)
}

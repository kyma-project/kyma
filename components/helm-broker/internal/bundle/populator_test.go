package bundle_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle/automock"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger/spy"
)

func TestPopulatorInitHappyPath(t *testing.T) {
	// GIVEN
	ts := newTestSuite()
	defer ts.AssertExpectationsOnMock(t)

	fixBundleReader := strings.NewReader(fixBundleContent())

	ts.mockRepository.ExpectOnIndexReader(fixIndexContent())
	ts.mockRepository.ExpectOnBundleReader(fixBundleName(), fixBundleVersion(), fixBundleReader)
	ts.mockLoader.ExpectOnLoad(fixBundleReader, fixBundle(), fixCharts())
	ts.mockBundleUpserter.ExpectOnUpsert(fixBundle())
	ts.mockChartUpserter.ExpectOnUpsert(fixChart())

	logSink := spy.NewLogSink()

	populator := bundle.NewPopulator(ts.mockRepository, ts.mockLoader, ts.mockBundleUpserter, ts.mockChartUpserter, logSink.Logger)
	// WHEN
	err := populator.Init()
	// THEN
	assert.NoError(t, err)

	logSink.AssertLogged(t, logrus.InfoLevel, "Bundle with name [meme] and version [0.10.0] successfully stored")
}

func TestPopulatorOnGetIndexFileError(t *testing.T) {
	// GIVEN
	ts := newTestSuite()
	defer ts.AssertExpectationsOnMock(t)

	ts.mockRepository.ExpectErrorOnIndexReader(fixError())

	logSink := spy.NewLogSink()

	populator := bundle.NewPopulator(ts.mockRepository, ts.mockLoader, ts.mockBundleUpserter, ts.mockChartUpserter, logSink.Logger)
	// WHEN
	err := populator.Init()
	// THEN
	assert.EqualError(t, err, fmt.Sprintf("while getting index file: %v", fixError()))
	assert.Empty(t, logSink.DumpAll())
}

func TestPopulatorOnGetBundleArchiveError(t *testing.T) {
	// GIVEN
	ts := newTestSuite()
	defer ts.AssertExpectationsOnMock(t)

	fixBundleReader := strings.NewReader(fixBundleContent())

	ts.mockRepository.ExpectOnIndexReader(fixIndexContent2Entries())
	ts.mockRepository.ExpectErrorOnBundleReader(fixBundleName(), fixBundleVersion(), fixError())
	ts.mockRepository.ExpectOnBundleReader("quote", fixBundleVersion(), fixBundleReader)
	ts.mockLoader.ExpectOnLoad(fixBundleReader, fixBundle(), fixCharts())
	ts.mockBundleUpserter.ExpectOnUpsert(fixBundle())
	ts.mockChartUpserter.ExpectOnUpsert(fixChart())

	logSink := spy.NewLogSink()

	populator := bundle.NewPopulator(ts.mockRepository, ts.mockLoader, ts.mockBundleUpserter, ts.mockChartUpserter, logSink.Logger)
	// WHEN
	err := populator.Init()
	// THEN
	require.NoError(t, err)
}

func TestPopulatorOnLoadingError(t *testing.T) {
	// GIVEN
	ts := newTestSuite()
	defer ts.AssertExpectationsOnMock(t)

	fixBundleReader := strings.NewReader(fixBundleContent())

	ts.mockRepository.ExpectOnIndexReader(fixIndexContent2Entries())
	ts.mockRepository.ExpectOnBundleReader("quote", fixBundleVersion(), fixBundleReader)
	ts.mockRepository.ExpectOnBundleReader(fixBundleName(), fixBundleVersion(), fixBundleReader)
	ts.mockLoader.ExpectErrorOnLoad(fixBundleReader, fixError())
	ts.mockLoader.ExpectOnLoad(fixBundleReader, fixBundle(), fixCharts())
	ts.mockBundleUpserter.ExpectOnUpsert(fixBundle())
	ts.mockChartUpserter.ExpectOnUpsert(fixChart())

	logSink := spy.NewLogSink()

	populator := bundle.NewPopulator(ts.mockRepository, ts.mockLoader, ts.mockBundleUpserter, ts.mockChartUpserter, logSink.Logger)
	// WHEN
	err := populator.Init()

	// THEN
	require.NoError(t, err)
}

func TestPopulatorOnUpsertingChartError(t *testing.T) {
	// GIVEN
	ts := newTestSuite()
	defer ts.AssertExpectationsOnMock(t)

	fixBundleReader := strings.NewReader(fixBundleContent())

	ts.mockRepository.ExpectOnIndexReader(fixIndexContent())
	ts.mockRepository.ExpectOnBundleReader(fixBundleName(), fixBundleVersion(), fixBundleReader)
	ts.mockLoader.ExpectOnLoad(fixBundleReader, fixBundle(), fixCharts())
	ts.mockChartUpserter.ExpectErrorOnUpsert(fixChart(), fixError())

	logSink := spy.NewLogSink()

	populator := bundle.NewPopulator(ts.mockRepository, ts.mockLoader, ts.mockBundleUpserter, ts.mockChartUpserter, logSink.Logger)
	// WHEN
	err := populator.Init()
	// THEN
	require.NoError(t, err)
}

func TestPopulatorOnUpsertingBundleError(t *testing.T) {
	// GIVEN
	ts := newTestSuite()
	defer ts.AssertExpectationsOnMock(t)

	fixBundleReader := strings.NewReader(fixBundleContent())

	ts.mockRepository.ExpectOnIndexReader(fixIndexContent())
	ts.mockRepository.ExpectOnBundleReader(fixBundleName(), fixBundleVersion(), fixBundleReader)
	ts.mockLoader.ExpectOnLoad(fixBundleReader, fixBundle(), fixCharts())
	ts.mockChartUpserter.ExpectOnUpsert(fixChart())
	ts.mockBundleUpserter.ExpectErrorOnUpsert(fixBundle(), fixError())

	logSink := spy.NewLogSink()

	populator := bundle.NewPopulator(ts.mockRepository, ts.mockLoader, ts.mockBundleUpserter, ts.mockChartUpserter, logSink.Logger)
	// WHEN
	err := populator.Init()
	// THEN
	require.NoError(t, err)
}

func fixError() error {
	return errors.New("some error")
}

func fixBundleName() bundle.Name {
	return "meme"
}

func fixBundleVersion() bundle.Version {
	return "0.10.0"
}

func fixBundleContent() string {
	return "data"
}

func fixBundle() *internal.Bundle {
	return &internal.Bundle{}
}

func fixCharts() []*chart.Chart {
	return []*chart.Chart{fixChart()}
}

func fixChart() *chart.Chart {
	return &chart.Chart{
		Values: &chart.Config{
			Raw: "value",
		},
	}
}

func fixIndexContent() string {
	return `
apiVersion: v1
entries:
  meme:
    - name: meme
      created: 2016-10-06T16:23:20.499814565-06:00
      description: Meme service
      version: 0.10.0
`
}

func fixIndexContent2Entries() string {
	return `
apiVersion: v1
entries:
  quote:
    - name: quote
      created: 2016-10-06T16:23:20.499814565-06:00
      description: Quote service
      version: 0.10.0
  meme:
    - name: meme
      created: 2016-10-06T16:23:20.499814565-06:00
      description: Meme service
      version: 0.10.0
`
}

type testSuite struct {
	mockRepository     *automock.Repository
	mockLoader         *automock.BundleLoader
	mockBundleUpserter *automock.BundleUpserter
	mockChartUpserter  *automock.ChartUpserter
}

func newTestSuite() testSuite {
	return testSuite{
		mockRepository:     &automock.Repository{},
		mockLoader:         &automock.BundleLoader{},
		mockBundleUpserter: &automock.BundleUpserter{},
		mockChartUpserter:  &automock.ChartUpserter{},
	}

}

func (ts *testSuite) AssertExpectationsOnMock(t *testing.T) {
	t.Helper()
	ts.mockRepository.AssertExpectations(t)
	ts.mockLoader.AssertExpectations(t)
	ts.mockBundleUpserter.AssertExpectations(t)
	ts.mockChartUpserter.AssertExpectations(t)

}

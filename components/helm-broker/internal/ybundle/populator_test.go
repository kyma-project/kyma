package ybundle_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/ybundle"
	"github.com/kyma-project/kyma/components/helm-broker/internal/ybundle/automock"
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

	populator := ybundle.NewPopulator(ts.mockRepository, ts.mockLoader, ts.mockBundleUpserter, ts.mockChartUpserter, logSink.Logger)
	// WHEN
	err := populator.Init()
	// THEN
	assert.NoError(t, err)

	logSink.AssertLogged(t, logrus.InfoLevel, "Bundle with name [meme] and version [0.10.0] successfully stored")

}

func TestPopulatorErrorOnGetIndexFile(t *testing.T) {
	// GIVEN
	ts := newTestSuite()
	defer ts.AssertExpectationsOnMock(t)

	ts.mockRepository.ExpectErrorOnIndexReader(fixError())

	logSink := spy.NewLogSink()

	populator := ybundle.NewPopulator(ts.mockRepository, ts.mockLoader, ts.mockBundleUpserter, ts.mockChartUpserter, logSink.Logger)
	// WHEN
	err := populator.Init()
	// THEN
	assert.EqualError(t, err, fmt.Sprintf("while getting index.yaml: %v", fixError()))
	assert.Empty(t, logSink.DumpAll())
}

func TestPopulatorErrorOnGetBundleArchive(t *testing.T) {
	// GIVEN
	ts := newTestSuite()
	defer ts.AssertExpectationsOnMock(t)

	ts.mockRepository.ExpectOnIndexReader(fixIndexContent())
	ts.mockRepository.ExpectErrorOnBundleReader(fixError())

	logSink := spy.NewLogSink()

	populator := ybundle.NewPopulator(ts.mockRepository, ts.mockLoader, ts.mockBundleUpserter, ts.mockChartUpserter, logSink.Logger)
	// WHEN
	err := populator.Init()
	// THEN
	assert.EqualError(t, err, fmt.Sprintf("while reading bundle archive for name [meme] and version [0.10.0]: %v", fixError()))
	assert.Empty(t, logSink.DumpAll())
}

func TestPopulatorErrorOnLoading(t *testing.T) {
	// GIVEN
	ts := newTestSuite()
	defer ts.AssertExpectationsOnMock(t)

	fixBundleReader := strings.NewReader(fixBundleContent())

	ts.mockRepository.ExpectOnIndexReader(fixIndexContent())
	ts.mockRepository.ExpectOnBundleReader(fixBundleName(), fixBundleVersion(), fixBundleReader)
	ts.mockLoader.ExpectErrorOnLoad(fixError())

	logSink := spy.NewLogSink()

	populator := ybundle.NewPopulator(ts.mockRepository, ts.mockLoader, ts.mockBundleUpserter, ts.mockChartUpserter, logSink.Logger)
	// WHEN
	err := populator.Init()
	// THEN
	assert.EqualError(t, err, fmt.Sprintf("while loading bundle and charts for bundle [meme] and version [0.10.0]: %v", fixError()))
	assert.Empty(t, logSink.DumpAll())
}

func TestPopulatorErrorOnUpsertingChart(t *testing.T) {
	// GIVEN
	ts := newTestSuite()
	defer ts.AssertExpectationsOnMock(t)

	fixBundleReader := strings.NewReader(fixBundleContent())

	ts.mockRepository.ExpectOnIndexReader(fixIndexContent())
	ts.mockRepository.ExpectOnBundleReader(fixBundleName(), fixBundleVersion(), fixBundleReader)
	ts.mockLoader.ExpectOnLoad(fixBundleReader, fixBundle(), fixCharts())
	ts.mockChartUpserter.ExpectErrorOnUpsert(fixError())

	logSink := spy.NewLogSink()

	populator := ybundle.NewPopulator(ts.mockRepository, ts.mockLoader, ts.mockBundleUpserter, ts.mockChartUpserter, logSink.Logger)
	// WHEN
	err := populator.Init()
	// THEN
	assert.EqualError(t, err, fmt.Sprintf("while storing chart [values:<raw:\"value\" > ] for bundle [meme] with version [0.10.0]: %v", fixError()))
	assert.Empty(t, logSink.DumpAll())
}

func TestPopulatorErrorOnUpsertingBundle(t *testing.T) {
	// GIVEN
	ts := newTestSuite()
	defer ts.AssertExpectationsOnMock(t)

	fixBundleReader := strings.NewReader(fixBundleContent())

	ts.mockRepository.ExpectOnIndexReader(fixIndexContent())
	ts.mockRepository.ExpectOnBundleReader(fixBundleName(), fixBundleVersion(), fixBundleReader)
	ts.mockLoader.ExpectOnLoad(fixBundleReader, fixBundle(), fixCharts())
	ts.mockChartUpserter.ExpectOnUpsert(fixChart())
	ts.mockBundleUpserter.ExpectErrorOnUpsert(fixError())

	logSink := spy.NewLogSink()

	populator := ybundle.NewPopulator(ts.mockRepository, ts.mockLoader, ts.mockBundleUpserter, ts.mockChartUpserter, logSink.Logger)
	// WHEN
	err := populator.Init()
	// THEN
	assert.EqualError(t, err, fmt.Sprintf("while storing bundle [meme] with version [0.10.0]: %v", fixError()))
	assert.Empty(t, logSink.DumpAll())
}

func fixError() error {
	return errors.New("some error")
}

func fixBundleName() ybundle.BundleName {
	return "meme"
}

func fixBundleVersion() ybundle.BundleVersion {
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
